//go:build !linux && !windows

package loadproof

func openHandleCount() (uint64, bool) {
	return 0, false
}
