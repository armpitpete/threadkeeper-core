//go:build linux && amd64

package sourceadapter

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"github.com/armpitpete/threadkeeper-core/internal/escrow"
)

// Linux/amd64 renameat2 is syscall 316. RENAME_EXCHANGE (2) swaps the names
// atomically so a source file can change between separate Lstat and Open calls.
func exchangeSourcePaths(a, b string) error {
	aPtr, err := syscall.BytePtrFromString(a)
	if err != nil {
		return err
	}
	bPtr, err := syscall.BytePtrFromString(b)
	if err != nil {
		return err
	}
	const linuxRenameat2 = 316
	const atFDCWD = ^uintptr(99) // -100, AT_FDCWD
	const renameExchange = 2
	_, _, errno := syscall.RawSyscall6(linuxRenameat2,
		atFDCWD, uintptr(unsafe.Pointer(aPtr)),
		atFDCWD, uintptr(unsafe.Pointer(bPtr)),
		renameExchange, 0)
	runtime.KeepAlive(aPtr)
	runtime.KeepAlive(bPtr)
	if errno != 0 {
		return errno
	}
	return nil
}

func TestFileAdapterRejectsConcurrentSymlinkEscape(t *testing.T) {
	root, externalDir := t.TempDir(), t.TempDir()
	filename := filepath.Join(root, "mutable.txt")
	alternate := filepath.Join(root, "alternate.txt")
	external := filepath.Join(externalDir, "external.txt")
	if err := os.WriteFile(filename, []byte("inside"), 0o600); err != nil {
		t.Fatal(err)
	}
	outside := []byte("outside: harmless fixture")
	if err := os.WriteFile(external, outside, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, alternate); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	adapter, err := NewFileAdapter(root)
	if err != nil {
		t.Fatal(err)
	}
	var stop atomic.Bool
	var swaps atomic.Int64
	swapErr := make(chan error, 1)
	go func() {
		for !stop.Load() {
			if err := exchangeSourcePaths(filename, alternate); err != nil {
				swapErr <- err
				return
			}
			swaps.Add(1)
		}
		swapErr <- nil
	}()
	request := Request{
		SourceID: "sample",
		RelativePath: "mutable.txt",
		ExpectedSHA256: escrow.HashContent(outside),
	}
	for i, start := 0, time.Now(); i < 40000 && time.Since(start) < 2*time.Second; i++ {
		snapshot, err := adapter.Fetch(context.Background(), request)
		if err == nil && string(snapshot.Content) == string(outside) {
			stop.Store(true)
			<-swapErr
			t.Fatalf("outside-root contents accepted after %d swaps", swaps.Load())
		}
	}
	stop.Store(true)
	if err := <-swapErr; err != nil {
		t.Skipf("atomic path-exchange unavailable: %v", err)
	}
	if swaps.Load() == 0 {
		t.Fatal("adversarial path swaps did not execute")
	}
}
