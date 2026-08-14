package quarantine

import (
	"errors"
	"os"
	"path/filepath"
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
