package ledger

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash"

	"github.com/armpitpete/threadkeeper-core/internal/quarantine"
)

const quarantineBindingPrefix = "bound-"
const quarantineStagePrefix = "stage-"

func rawQuarantineIdentity(raw []byte) (string, int64) {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), int64(len(raw))
}

// quarantineStageID gives each Prepare invocation exclusive cleanup ownership
// of its temporary staging file. Staging identities are deliberately not
// content-derived: identical concurrent prepares must not be able to remove one
// another's in-flight staging material. Randomness failure returns an invalid
// empty ID, which the quarantine store rejects fail-closed.
func quarantineStageID(_ []byte) string {
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return ""
	}
	return quarantineStagePrefix + hex.EncodeToString(nonce[:])
}

// candidateQuarantineBindingID binds the private quarantine capability to the
// complete prepared candidate identity. The file stored under this ID contains
// the exact raw event bytes. A caller can recompute another ID for a forged
// commit/path, but preparation will not have materialised a file at that ID.
func candidateQuarantineBindingID(candidate WriteCandidate) (string, error) {
	return candidateQuarantineBindingIDFields(
		candidate.ExpectedHead,
		candidate.CandidateCommit,
		candidate.EventPath,
		candidate.EventID,
		candidate.IdempotencyKey,
		candidate.ContentSHA256,
		candidate.Quarantine.ContentSHA256,
		candidate.Quarantine.Size,
	)
}

func candidateQuarantineBindingIDFields(expectedHead, candidateCommit, eventPath, eventID, idempotencyKey, contentSHA256, rawSHA256 string, rawSize int64) (string, error) {
	if expectedHead == "" || candidateCommit == "" || eventPath == "" || eventID == "" || idempotencyKey == "" || contentSHA256 == "" || rawSHA256 == "" || rawSize <= 0 {
		return "", fmt.Errorf("CANDIDATE_INVALID: incomplete quarantine binding identity")
	}
	h := sha256.New()
	for _, field := range []string{
		expectedHead,
		candidateCommit,
		eventPath,
		eventID,
		idempotencyKey,
		contentSHA256,
		rawSHA256,
	} {
		writeBindingField(h, []byte(field))
	}
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(rawSize))
	writeBindingField(h, size[:])
	return quarantineBindingPrefix + hex.EncodeToString(h.Sum(nil)), nil
}

func writeBindingField(h hash.Hash, field []byte) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(field)))
	_, _ = h.Write(length[:])
	_, _ = h.Write(field)
}

func expectedBoundQuarantineEntry(candidate WriteCandidate) (quarantine.Entry, error) {
	id, err := candidateQuarantineBindingID(candidate)
	if err != nil {
		return quarantine.Entry{}, err
	}
	return quarantine.Entry{
		ID:            id,
		ContentSHA256: candidate.Quarantine.ContentSHA256,
		Size:          candidate.Quarantine.Size,
	}, nil
}
