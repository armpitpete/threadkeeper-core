package gitledger

import "context"

// ReadRegularFile reads one exact tree path only after requiring the accepted
// durable representation to be an ordinary 100644 blob. Semantic authority
// records such as Genesis must use this instead of treating a symlink or
// executable-mode blob as equivalent content.
func (r *Reader) ReadRegularFile(ctx context.Context, commit, path string) ([]byte, error) {
	if err := r.requireRegularBlobAt(ctx, commit, path); err != nil {
		return nil, err
	}
	return r.ReadFile(ctx, commit, path)
}
