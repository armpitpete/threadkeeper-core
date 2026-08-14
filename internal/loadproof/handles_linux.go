//go:build linux

package loadproof

import "os"

func openHandleCount() (uint64, bool) {
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		return 0, false
	}
	return uint64(len(entries)), true
}
