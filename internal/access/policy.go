package access

import (
	"fmt"
	"time"
)

type Classification string

const (
	Public       Classification = "public"
	Internal     Classification = "internal"
	Confidential Classification = "confidential"
	Restricted   Classification = "restricted"
)

var rank = map[Classification]int{Public: 0, Internal: 1, Confidential: 2, Restricted: 3}

func (c Classification) Validate() error {
	if _, ok := rank[c]; !ok { return fmt.Errorf("CLASSIFICATION_INVALID: %q", c) }
	return nil
}

func CanRead(clearance, record Classification) bool {
	cr, cok := rank[clearance]
	rr, rok := rank[record]
	return cok && rok && cr >= rr
}

type RedactionTombstone struct {
	RecordID           string `json:"record_id"`
	AuthorisedBy       string `json:"authorised_by"`
	Reason             string `json:"reason"`
	ContentDestroyedAt string `json:"content_destroyed_at"`
}

func (r RedactionTombstone) Validate() error {
	if r.RecordID == "" || r.AuthorisedBy == "" || r.Reason == "" { return fmt.Errorf("REDACTION_INVALID: identity, authority and reason are required") }
	if _, err := time.Parse(time.RFC3339, r.ContentDestroyedAt); err != nil { return fmt.Errorf("REDACTION_INVALID: content_destroyed_at: %w", err) }
	return nil
}
