package digest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/armpitpete/threadkeeper-core/internal/canonicaljson"
	"github.com/armpitpete/threadkeeper-core/internal/strictjson"
)

const Field = "content_sha256"

var ErrRootNotObject = errors.New("durable record root must be an object")

func Compute(raw []byte) (string, []byte, error) {
	v, err := strictjson.Decode(raw)
	if err != nil {
		return "", nil, err
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return "", nil, ErrRootNotObject
	}
	delete(obj, Field)
	payload, err := json.Marshal(obj)
	if err != nil {
		return "", nil, fmt.Errorf("marshal digest payload: %w", err)
	}
	canonical, err := canonicaljson.Canonicalize(payload)
	if err != nil {
		return "", nil, err
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), canonical, nil
}

func Complete(raw []byte) ([]byte, string, error) {
	v, err := strictjson.Decode(raw)
	if err != nil {
		return nil, "", err
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return nil, "", ErrRootNotObject
	}
	d, _, err := Compute(raw)
	if err != nil {
		return nil, "", err
	}
	obj[Field] = d
	stored, err := json.Marshal(obj)
	if err != nil {
		return nil, "", fmt.Errorf("marshal completed record: %w", err)
	}
	canonical, err := canonicaljson.Canonicalize(stored)
	if err != nil {
		return nil, "", err
	}
	return canonical, d, nil
}

func Verify(raw []byte) error {
	v, err := strictjson.Decode(raw)
	if err != nil {
		return err
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return ErrRootNotObject
	}
	stored, ok := obj[Field].(string)
	if !ok || stored == "" {
		return fmt.Errorf("%s missing or not a string", Field)
	}
	actual, _, err := Compute(raw)
	if err != nil {
		return err
	}
	if stored != actual {
		return fmt.Errorf("DIGEST_MISMATCH: stored %s actual %s", stored, actual)
	}
	return nil
}
