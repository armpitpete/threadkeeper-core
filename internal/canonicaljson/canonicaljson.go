package canonicaljson

import (
	"fmt"

	"github.com/armpitpete/threadkeeper-core/internal/strictjson"
	"github.com/gowebpki/jcs"
)

func Canonicalize(raw []byte) ([]byte, error) {
	if err := strictjson.Validate(raw); err != nil {
		return nil, err
	}
	out, err := jcs.Transform(raw)
	if err != nil {
		return nil, fmt.Errorf("JCS canonicalization failed: %w", err)
	}
	return out, nil
}
