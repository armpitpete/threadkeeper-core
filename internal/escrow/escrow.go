package escrow

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type Mode string

const (
	ReferenceOnly          Mode = "reference_only"
	MetadataSnapshot       Mode = "metadata_snapshot"
	ContentSnapshot        Mode = "content_snapshot"
	ExternallyDurable      Mode = "externally_durable"
	PreservationProhibited Mode = "preservation_prohibited"
)

type Policy struct {
	Mode     Mode  `json:"mode"`
	MaxBytes int64 `json:"max_bytes,omitempty"`
}

type Snapshot struct {
	SourceID      string `json:"source_id"`
	VersionID     string `json:"version_id"`
	ContentSHA256 string `json:"content_sha256"`
	Size          int64  `json:"size"`
	MediaType     string `json:"media_type,omitempty"`
}

func (p Policy) Validate() error {
	switch p.Mode {
	case ReferenceOnly, MetadataSnapshot, ContentSnapshot, ExternallyDurable, PreservationProhibited:
	default:
		return fmt.Errorf("ESCROW_POLICY_INVALID: unknown mode %q", p.Mode)
	}
	if p.MaxBytes < 0 { return fmt.Errorf("ESCROW_POLICY_INVALID: max_bytes must be non-negative") }
	if p.Mode == PreservationProhibited && p.MaxBytes != 0 { return fmt.Errorf("ESCROW_POLICY_INVALID: prohibited preservation cannot allocate content bytes") }
	return nil
}

func HashContent(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func VerifyContent(snapshot Snapshot, content []byte) error {
	if snapshot.SourceID == "" || snapshot.VersionID == "" || snapshot.ContentSHA256 == "" {
		return fmt.Errorf("ESCROW_INVALID: source_id, version_id and content_sha256 are required")
	}
	if int64(len(content)) != snapshot.Size { return fmt.Errorf("ESCROW_SIZE_MISMATCH: got %d want %d", len(content), snapshot.Size) }
	if got := HashContent(content); got != snapshot.ContentSHA256 { return fmt.Errorf("ESCROW_DIGEST_MISMATCH: got %s want %s", got, snapshot.ContentSHA256) }
	return nil
}
