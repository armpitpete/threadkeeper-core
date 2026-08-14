package quarantine

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestEnsureFailedCreatorCannotInvalidateConcurrentSuccess(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	id := "concurrent-identical"
	content := []byte("complete identical candidate bytes")
	beforeSync := make(chan struct{})
	releaseSync := make(chan struct{})
	creatorDone := make(chan error, 1)

	go func() {
		_, err := s.ensure(id, content, &putHooks{
			beforeSync: func() {
				close(beforeSync)
				<-releaseSync
			},
			syncFile: func(*os.File) error {
				return errors.New("injected sync failure")
			},
		})
		creatorDone <- err
	}()

	<-beforeSync
	finalPath := filepath.Join(s.dir, candidateName(id))
	if _, err := os.Stat(finalPath); !os.IsNotExist(err) {
		t.Fatalf("uncompleted creator exposed final capability before sync/close: %v", err)
	}

	// The second identical publisher must succeed through its own fully durable
	// private publication; it must not depend on bytes owned by the blocked first
	// creator.
	second, err := s.Ensure(id, content)
	if err != nil {
		t.Fatalf("independent identical publisher failed: %v", err)
	}
	want := entryFor(id, content)
	if second != want {
		t.Fatalf("second publication = %#v want %#v", second, want)
	}

	close(releaseSync)
	if err := <-creatorDone; err == nil || err.Error() != "injected sync failure" {
		t.Fatalf("first creator error = %v, want injected sync failure", err)
	}

	gotEntry, got, err := s.ReadID(id)
	if err != nil {
		t.Fatalf("failed creator removed another publisher's final capability: %v", err)
	}
	if gotEntry != want || string(got) != string(content) {
		t.Fatalf("surviving final capability entry=%#v bytes=%q", gotEntry, got)
	}

	temps, err := filepath.Glob(filepath.Join(s.dir, publicationTempPrefix+"*"+publicationTempSuffix))
	if err != nil {
		t.Fatal(err)
	}
	if len(temps) != 0 {
		t.Fatalf("private publication files leaked after failure/success race: %v", temps)
	}
}

func TestEnsureVisibleCandidateRequiresSuccessfulDirectorySync(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	id := "visible-before-directory-sync"
	content := []byte("directory durability matters")
	beforeDirectorySync := make(chan struct{})
	releaseCreator := make(chan struct{})
	creatorDone := make(chan error, 1)

	go func() {
		_, err := s.ensure(id, content, &putHooks{
			beforeDirectorySync: func() {
				close(beforeDirectorySync)
				<-releaseCreator
			},
			syncDirectory: func(*os.File) error {
				return errors.New("injected creator directory sync failure")
			},
		})
		creatorDone <- err
	}()

	<-beforeDirectorySync
	finalPath := filepath.Join(s.dir, candidateName(id))
	if _, err := os.Stat(finalPath); err != nil {
		t.Fatalf("expected final capability to be visible before creator directory sync: %v", err)
	}

	// A second caller is allowed to converge on the visible identical file only
	// if it establishes directory durability itself before returning success.
	secondSyncs := 0
	second, err := s.ensure(id, content, &putHooks{
		syncDirectory: func(dir *os.File) error {
			secondSyncs++
			return dir.Sync()
		},
	})
	if err != nil {
		t.Fatalf("second publisher failed to durably converge: %v", err)
	}
	if secondSyncs == 0 {
		t.Fatal("second publisher reported success without syncing the quarantine directory")
	}
	want := entryFor(id, content)
	if second != want {
		t.Fatalf("second publication = %#v want %#v", second, want)
	}

	close(releaseCreator)
	if err := <-creatorDone; err == nil || !strings.Contains(err.Error(), "injected creator directory sync failure") {
		t.Fatalf("creator error = %v, want injected directory sync failure", err)
	}

	gotEntry, got, err := s.ReadID(id)
	if err != nil {
		t.Fatalf("failed creator invalidated second publisher's durable capability: %v", err)
	}
	if gotEntry != want || string(got) != string(content) {
		t.Fatalf("durable final capability entry=%#v bytes=%q", gotEntry, got)
	}
}

func TestEnsureDirectorySyncFailureNeverReportsSuccess(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	id := "directory-sync-failure"
	content := []byte("candidate may be visible but not yet durable")
	_, err = s.ensure(id, content, &putHooks{
		syncDirectory: func(*os.File) error {
			return errors.New("injected directory sync failure")
		},
	})
	if err == nil || !strings.Contains(err.Error(), "QUARANTINE_DURABILITY_FAILURE") {
		t.Fatalf("directory sync failure = %v, want durability failure", err)
	}

	// The failed call may leave a visible final hard link. A subsequent caller
	// must establish durability itself before it can recover idempotently.
	recovered, err := s.Ensure(id, content)
	if err != nil {
		t.Fatalf("subsequent durable recovery failed: %v", err)
	}
	want := entryFor(id, content)
	if recovered != want {
		t.Fatalf("recovered entry = %#v want %#v", recovered, want)
	}
}

func TestEnsureRecoversAfterProcessDeathBeforeDirectorySync(t *testing.T) {
	const helperEnv = "THREADKEEPER_QUARANTINE_CRASH_HELPER"
	const dirEnv = "THREADKEEPER_QUARANTINE_CRASH_DIR"
	id := "process-crash-before-directory-sync"
	content := []byte("process crash leaves only unconfirmed publication residue")

	if os.Getenv(helperEnv) == "1" {
		dir := os.Getenv(dirEnv)
		if dir == "" {
			os.Exit(91)
		}
		s, err := Open(dir)
		if err != nil {
			os.Exit(92)
		}
		// Simulate abrupt process death after the final hard link becomes visible
		// but before the correctness-critical directory Sync. os.Exit deliberately
		// bypasses deferred temp cleanup and Store.Close.
		_, _ = s.ensure(id, content, &putHooks{
			beforeDirectorySync: func() { os.Exit(0) },
		})
		os.Exit(93)
	}

	dir := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestEnsureRecoversAfterProcessDeathBeforeDirectorySync$")
	cmd.Env = append(os.Environ(), helperEnv+"=1", dirEnv+"="+dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("crash helper failed: %v\n%s", err, out)
	}

	finalPath := filepath.Join(dir, candidateName(id))
	if _, err := os.Stat(finalPath); err != nil {
		t.Fatalf("crash helper did not reach visible final publication: %v", err)
	}
	temps, err := filepath.Glob(filepath.Join(dir, publicationTempPrefix+"*"+publicationTempSuffix))
	if err != nil {
		t.Fatal(err)
	}
	if len(temps) == 0 {
		t.Fatal("expected abrupt process death to leave private publication residue")
	}

	// Restart with a fresh pinned Store. The visible final path is not trusted as
	// completed by itself: Ensure must execute the existing-final directory-sync
	// and revalidation path before returning the exact entry.
	restarted, err := OpenExisting(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Close()
	directorySyncs := 0
	recovered, err := restarted.ensure(id, content, &putHooks{
		syncDirectory: func(dir *os.File) error {
			directorySyncs++
			return dir.Sync()
		},
	})
	if err != nil {
		t.Fatalf("restart failed to establish durable publication: %v", err)
	}
	if directorySyncs == 0 {
		t.Fatal("restart accepted crash residue without establishing directory durability")
	}
	want := entryFor(id, content)
	if recovered != want {
		t.Fatalf("restart recovered entry = %#v want %#v", recovered, want)
	}
	gotEntry, got, err := restarted.ReadID(id)
	if err != nil {
		t.Fatal(err)
	}
	if gotEntry != want || string(got) != string(content) {
		t.Fatalf("restart final capability entry=%#v bytes=%q", gotEntry, got)
	}
}

func TestEnsureConcurrentIdenticalPublishersConvergeOnOneFinalCapability(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	id := "concurrent-identical-success"
	content := []byte("same complete candidate bytes")
	start := make(chan struct{})
	type result struct {
		entry Entry
		err   error
	}
	results := make(chan result, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			entry, err := s.Ensure(id, content)
			results <- result{entry: entry, err: err}
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	want := entryFor(id, content)
	for got := range results {
		if got.err != nil {
			t.Fatalf("identical concurrent publication failed: %v", got.err)
		}
		if got.entry != want {
			t.Fatalf("converged entry = %#v want %#v", got.entry, want)
		}
	}

	gotEntry, got, err := s.ReadID(id)
	if err != nil {
		t.Fatal(err)
	}
	if gotEntry != want || string(got) != string(content) {
		t.Fatalf("final capability entry=%#v bytes=%q", gotEntry, got)
	}
	temps, err := filepath.Glob(filepath.Join(s.dir, publicationTempPrefix+"*"+publicationTempSuffix))
	if err != nil {
		t.Fatal(err)
	}
	if len(temps) != 0 {
		t.Fatalf("private publication files leaked: %v", temps)
	}
}
