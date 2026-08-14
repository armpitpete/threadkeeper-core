package quarantine

import (
	"errors"
	"os"
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
