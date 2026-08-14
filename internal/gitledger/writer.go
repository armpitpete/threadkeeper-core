package gitledger

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrStaleState             = errors.New("STALE_STATE")
	ErrCandidateNotChild      = errors.New("CANDIDATE_NOT_EXACT_CHILD")
	ErrCASOutcomeUnknown      = errors.New("POST_CAS_RECOVERY_REQUIRED")
	ErrPostCASVerification    = errors.New("POST_CAS_VERIFICATION_FAILED")
	ErrCASAcceptanceRecovered = fmt.Errorf("%w: CAS_ACCEPTANCE_RECOVERED", ErrStaleState)
)

const (
	casRecoveryTimeout          = 30 * time.Second
	casContentionSettleTimeout  = time.Second
	casContentionRetryInterval  = 10 * time.Millisecond
)

type CandidateCommit struct {
	ExpectedHead string
	Commit       string
	Tree         string
	EventPath    string
	EventBlob    string
}

type recoveredCASState struct {
	Head               string
	CandidateInHistory bool
}

// Internal deterministic instrumentation for hostile CAS race tests.
// Production callers always use nil hooks.
type compareAndSwapHooks struct {
	afterInitialUpdateErrorRecovery func(recoveredCASState)
}

// PrepareEventCommit creates Git objects for exactly one new durable event
// without changing any ref. The candidate is not authoritative until a later
// successful CompareAndSwap.
func (r *Reader) PrepareEventCommit(ctx context.Context, expectedHead, eventPath string, event []byte, eventID string) (*CandidateCommit, error) {
	if !isObjectID(expectedHead) {
		return nil, fmt.Errorf("invalid expected head %q", expectedHead)
	}
	expectedHead = strings.ToLower(expectedHead)
	if err := validateEventPath(eventPath); err != nil {
		return nil, err
	}
	if eventID == "" || strings.ContainsAny(eventID, "\x00\r\n") {
		return nil, fmt.Errorf("invalid event id")
	}
	if err := r.CheckHistorySafety(ctx); err != nil {
		return nil, err
	}

	current, err := r.Head(ctx)
	if err != nil {
		return nil, err
	}
	if current != expectedHead {
		return nil, fmt.Errorf("%w: expected %s current %s", ErrStaleState, expectedHead, current)
	}

	out, err := r.run(ctx, "ls-tree", "-z", expectedHead, "--", eventPath)
	if err != nil {
		return nil, err
	}
	if len(out) != 0 {
		return nil, fmt.Errorf("EVENT_PATH_EXISTS: %s", eventPath)
	}

	indexDir, err := os.MkdirTemp("", "threadkeeper-index-*")
	if err != nil {
		return nil, fmt.Errorf("create private Git index directory: %w", err)
	}
	defer os.RemoveAll(indexDir)
	indexPath := filepath.Join(indexDir, "index")
	hooksDir, err := os.MkdirTemp("", "threadkeeper-no-hooks-*")
	if err != nil {
		return nil, fmt.Errorf("create empty hooks directory: %w", err)
	}
	defer os.RemoveAll(hooksDir)

	extra := []string{"GIT_INDEX_FILE=" + indexPath}
	if _, err := r.runWrite(ctx, nil, extra, hooksDir, "read-tree", expectedHead); err != nil {
		return nil, err
	}
	blobOut, err := r.runWrite(ctx, event, nil, hooksDir, "hash-object", "-w", "--stdin")
	if err != nil {
		return nil, err
	}
	blob := strings.TrimSpace(string(blobOut))
	if !isObjectID(blob) {
		return nil, fmt.Errorf("GIT_FAILURE: invalid event blob id %q", blob)
	}
	if _, err := r.runWrite(ctx, nil, extra, hooksDir, "update-index", "--add", "--cacheinfo", "100644", blob, eventPath); err != nil {
		return nil, err
	}
	treeOut, err := r.runWrite(ctx, nil, extra, hooksDir, "write-tree")
	if err != nil {
		return nil, err
	}
	tree := strings.TrimSpace(string(treeOut))
	if !isObjectID(tree) {
		return nil, fmt.Errorf("GIT_FAILURE: invalid candidate tree id %q", tree)
	}

	message := []byte("Threadkeeper event " + eventID + "\n")
	commitOut, err := r.runWrite(ctx, message, []string{
		"GIT_AUTHOR_NAME=Threadkeeper Core",
		"GIT_AUTHOR_EMAIL=threadkeeper-core@localhost",
		"GIT_AUTHOR_DATE=2000-01-01T00:00:00Z",
		"GIT_COMMITTER_NAME=Threadkeeper Core",
		"GIT_COMMITTER_EMAIL=threadkeeper-core@localhost",
		"GIT_COMMITTER_DATE=2000-01-01T00:00:00Z",
	}, hooksDir, "commit-tree", tree, "-p", expectedHead)
	if err != nil {
		return nil, err
	}
	commit := strings.TrimSpace(string(commitOut))
	if !isObjectID(commit) {
		return nil, fmt.Errorf("GIT_FAILURE: invalid candidate commit id %q", commit)
	}
	if err := r.VerifyEventCandidate(ctx, expectedHead, commit, eventPath); err != nil {
		return nil, err
	}
	return &CandidateCommit{
		ExpectedHead: expectedHead,
		Commit:       strings.ToLower(commit),
		Tree:         strings.ToLower(tree),
		EventPath:    eventPath,
		EventBlob:    strings.ToLower(blob),
	}, nil
}

// VerifyEventCandidate rechecks that a candidate is exactly one safe event
// addition on top of the expected head. Acceptance callers must invoke this
// rather than trusting a serialized or caller-supplied candidate handle.
func (r *Reader) VerifyEventCandidate(ctx context.Context, expectedHead, candidateCommit, eventPath string) error {
	if !isObjectID(expectedHead) || !isObjectID(candidateCommit) {
		return fmt.Errorf("invalid candidate object id")
	}
	if err := validateEventPath(eventPath); err != nil {
		return err
	}
	if err := r.CheckHistorySafety(ctx); err != nil {
		return err
	}
	if err := r.verifyExactChild(ctx, strings.ToLower(expectedHead), strings.ToLower(candidateCommit)); err != nil {
		return err
	}
	return r.verifySinglePathAddition(ctx, strings.ToLower(candidateCommit), eventPath)
}

// CompareAndSwap advances the authoritative ref only if it still equals the
// expected head. A candidate must be the exact single-parent child of that
// head. Repository hooks are bypassed so they cannot become hidden authority.
//
// Once update-ref is invoked, outcome recovery no longer depends on the caller
// context: that context may be cancelled after Git has already moved authority.
func (r *Reader) CompareAndSwap(ctx context.Context, expectedHead, candidateCommit string) error {
	return r.compareAndSwap(ctx, expectedHead, candidateCommit, nil)
}

func (r *Reader) compareAndSwap(ctx context.Context, expectedHead, candidateCommit string, hooks *compareAndSwapHooks) error {
	if !isObjectID(expectedHead) || !isObjectID(candidateCommit) {
		return fmt.Errorf("invalid compare-and-swap object id")
	}
	expectedHead = strings.ToLower(expectedHead)
	candidateCommit = strings.ToLower(candidateCommit)
	if err := r.CheckHistorySafety(ctx); err != nil {
		return err
	}
	if err := r.verifyExactChild(ctx, expectedHead, candidateCommit); err != nil {
		return err
	}

	current, err := r.Head(ctx)
	if err != nil {
		return err
	}
	if current != expectedHead {
		return fmt.Errorf("%w: expected %s current %s", ErrStaleState, expectedHead, current)
	}
	// Recheck repository-local safety immediately before the authority-changing
	// ref operation. The runtime deployment must additionally make the ledger
	// directory service-owned so untrusted processes cannot race this check.
	if err := r.CheckHistorySafety(ctx); err != nil {
		return err
	}

	hooksDir, err := os.MkdirTemp("", "threadkeeper-no-hooks-*")
	if err != nil {
		return fmt.Errorf("create empty hooks directory: %w", err)
	}
	defer os.RemoveAll(hooksDir)
	// --no-deref makes the CAS target the configured ref object itself. A
	// symbolic ref introduced by a filesystem race therefore cannot redirect
	// mutation to its target; static symbolic refs are rejected before this
	// point by checkAuthoritativeRefSafety.
	_, updateErr := r.runWrite(ctx, nil, nil, hooksDir, "update-ref", "--no-deref", r.ref, candidateCommit, expectedHead)

	// From this point onward the caller context is not evidence about whether
	// authority moved. Resolve the authoritative ref and, when it has advanced
	// beyond H1, inspect the recovered linear history using a fresh bounded
	// context. This distinguishes "another writer won H0" from "our H1 was
	// accepted and a later valid writer already advanced to H2".
	recovered, recoveryErr := r.recoverStateAfterCAS(candidateCommit)
	if updateErr != nil {
		if recoveryErr != nil {
			return fmt.Errorf("%w: update-ref returned %v; authoritative-state recovery failed: %v", ErrCASOutcomeUnknown, updateErr, recoveryErr)
		}
		if recovered.CandidateInHistory {
			// H1 is durably authoritative or in the current authoritative history,
			// but because update-ref itself returned an error this invocation cannot
			// prove whether it performed the transition or an identical concurrent
			// writer did. Return an explicit recovered-acceptance condition which
			// also unwraps to STALE_STATE so the ledger layer reconstructs the
			// durable idempotent result as already_accepted.
			return fmt.Errorf("%w: candidate %s is present in recovered authoritative history after update-ref error: %v", ErrCASAcceptanceRecovered, candidateCommit, updateErr)
		}
		if recovered.Head == expectedHead {
			// A failed update-ref while H0 is still visible is not proof that no
			// identical writer is currently holding the ref lock and will publish H1
			// immediately afterwards. Give that ambiguity a detached bounded settle
			// window; if it still cannot be resolved, return explicit unknown rather
			// than an ordinary write failure.
			if hooks != nil && hooks.afterInitialUpdateErrorRecovery != nil {
				hooks.afterInitialUpdateErrorRecovery(recovered)
			}
			settled, settleErr := r.settleStateAfterCASContention(expectedHead, candidateCommit)
			if settleErr != nil {
				return fmt.Errorf("%w: update-ref returned %v; contention recovery failed: %v", ErrCASOutcomeUnknown, updateErr, settleErr)
			}
			if settled.CandidateInHistory {
				return fmt.Errorf("%w: candidate %s appeared in authoritative history during contention recovery after update-ref error: %v", ErrCASAcceptanceRecovered, candidateCommit, updateErr)
			}
			if settled.Head != expectedHead {
				return fmt.Errorf("%w: expected %s current %s", ErrStaleState, expectedHead, settled.Head)
			}
			return fmt.Errorf("%w: update-ref returned %v; authoritative ref remained %s through contention recovery window", ErrCASOutcomeUnknown, updateErr, expectedHead)
		}
		// Another exact-head writer won H0 and H1 is absent from the resulting
		// authoritative history.
		return fmt.Errorf("%w: expected %s current %s", ErrStaleState, expectedHead, recovered.Head)
	}

	// A successful update-ref means H1 was authoritative at the atomic update.
	// A current descendant containing H1 is also a verified acceptance. Any
	// inability to prove H1 in the recovered authority history is an explicit
	// post-CAS recovery condition, never an ordinary write failure.
	if recoveryErr != nil {
		return fmt.Errorf("%w: update-ref accepted %s but verification failed: %v", ErrPostCASVerification, candidateCommit, recoveryErr)
	}
	if !recovered.CandidateInHistory {
		return fmt.Errorf("%w: update-ref accepted %s but recovered head %s does not contain it", ErrPostCASVerification, candidateCommit, recovered.Head)
	}
	return nil
}

func (r *Reader) recoverStateAfterCAS(candidateCommit string) (recoveredCASState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), casRecoveryTimeout)
	defer cancel()
	return r.recoverStateAfterCASContext(ctx, candidateCommit)
}

func (r *Reader) recoverStateAfterCASContext(ctx context.Context, candidateCommit string) (recoveredCASState, error) {
	if err := r.CheckHistorySafety(ctx); err != nil {
		return recoveredCASState{}, fmt.Errorf("repository safety check: %w", err)
	}
	got, err := r.Head(ctx)
	if err != nil {
		return recoveredCASState{}, err
	}
	state := recoveredCASState{Head: got, CandidateInHistory: got == candidateCommit}
	if state.CandidateInHistory {
		return state, nil
	}
	history, err := r.History(ctx, got)
	if err != nil {
		return recoveredCASState{}, err
	}
	for _, commit := range history {
		if commit.ID == candidateCommit {
			state.CandidateInHistory = true
			break
		}
	}
	return state, nil
}

// settleStateAfterCASContention resolves the narrow state in which update-ref
// returned an error but a detached recovery still saw H0. A legitimate writer
// may still own the Git ref lock and publish immediately afterwards. Poll only
// for a bounded interval; an unchanged H0 at the deadline remains ambiguous and
// is reported by the caller as POST_CAS_RECOVERY_REQUIRED.
func (r *Reader) settleStateAfterCASContention(expectedHead, candidateCommit string) (recoveredCASState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), casContentionSettleTimeout)
	defer cancel()
	ticker := time.NewTicker(casContentionRetryInterval)
	defer ticker.Stop()

	var last recoveredCASState
	for {
		state, err := r.recoverStateAfterCASContext(ctx, candidateCommit)
		if err != nil {
			if ctx.Err() != nil {
				return last, fmt.Errorf("contention recovery deadline: %w", ctx.Err())
			}
			return last, err
		}
		last = state
		if state.CandidateInHistory || state.Head != expectedHead {
			return state, nil
		}
		select {
		case <-ctx.Done():
			return last, nil
		case <-ticker.C:
		}
	}
}

func (r *Reader) verifyExactChild(ctx context.Context, expectedHead, candidateCommit string) error {
	out, err := r.run(ctx, "rev-list", "--parents", "-n", "1", candidateCommit)
	if err != nil {
		return err
	}
	fields := strings.Fields(strings.TrimSpace(string(out)))
	if len(fields) != 2 || strings.ToLower(fields[0]) != strings.ToLower(candidateCommit) || strings.ToLower(fields[1]) != strings.ToLower(expectedHead) {
		return fmt.Errorf("%w: candidate %s must have sole parent %s", ErrCandidateNotChild, candidateCommit, expectedHead)
	}
	return nil
}

func (r *Reader) verifySinglePathAddition(ctx context.Context, candidateCommit, eventPath string) error {
	out, err := r.run(ctx, "diff-tree", "--no-commit-id", "--name-status", "-r", "-z", candidateCommit)
	if err != nil {
		return err
	}
	tokens := bytes.Split(out, []byte{0})
	nonEmpty := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if len(token) != 0 {
			nonEmpty = append(nonEmpty, string(token))
		}
	}
	if len(nonEmpty) != 2 || nonEmpty[0] != "A" || nonEmpty[1] != eventPath {
		return fmt.Errorf("CANDIDATE_TREE_MISMATCH: candidate must add only %q, got %q", eventPath, nonEmpty)
	}
	return nil
}

func (r *Reader) runWrite(parent context.Context, stdin []byte, extraEnv []string, hooksDir string, args ...string) ([]byte, error) {
	releaseRoot, err := r.holdRepositoryRoot()
	if err != nil {
		return nil, err
	}
	defer releaseRoot()

	ctx, cancel := context.WithTimeout(parent, r.timeout)
	defer cancel()
	base := []string{"--no-replace-objects", "--git-dir=" + r.gitDir, "-c", "core.hooksPath=" + hooksDir, "-c", "commit.gpgSign=false"}
	cmd := exec.CommandContext(ctx, r.gitPath, append(base, args...)...)
	cmd.Env = controlledWriteEnv(extraEnv)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("GIT_FAILURE: write command timed out")
		}
		msg := strings.TrimSpace(stderr.String())
		if len(msg) > 800 {
			msg = msg[:800]
		}
		command := "git"
		if len(args) > 0 {
			command += " " + args[0]
		}
		return nil, fmt.Errorf("GIT_FAILURE: %s failed: %w: %s", command, err, msg)
	}
	return stdout.Bytes(), nil
}

func controlledWriteEnv(extra []string) []string {
	base := controlledEnv()
	env := make([]string, 0, len(base)+len(extra))
	for _, item := range base {
		if strings.HasPrefix(item, "GIT_AUTHOR_") || strings.HasPrefix(item, "GIT_COMMITTER_") || strings.HasPrefix(item, "EMAIL=") {
			continue
		}
		env = append(env, item)
	}
	return append(env, extra...)
}

func validateEventPath(p string) error {
	if !strings.HasPrefix(p, "events/") || !strings.HasSuffix(p, ".json") || p != path.Clean(p) || strings.Contains(p, "\\") || strings.ContainsAny(p, "\x00\r\n*?[") {
		return fmt.Errorf("invalid durable event path %q", p)
	}
	for _, part := range strings.Split(p, "/") {
		if part == "" || part == "." || part == ".." {
			return fmt.Errorf("invalid durable event path %q", p)
		}
		for _, c := range part {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.') {
				return fmt.Errorf("invalid durable event path %q", p)
			}
	}
	return nil
}
