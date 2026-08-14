//go:build windows

package loadproof

import (
	"syscall"
	"unsafe"
)

var (
	kernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procGetCurrentProcess     = kernel32.NewProc("GetCurrentProcess")
	procGetProcessHandleCount = kernel32.NewProc("GetProcessHandleCount")
)

func openHandleCount() (uint64, bool) {
	process, _, _ := procGetCurrentProcess.Call()
	var count uint32
	ok, _, _ := procGetProcessHandleCount.Call(process, uintptr(unsafe.Pointer(&count)))
	if ok == 0 {
		return 0, false
	}
	return uint64(count), true
}
