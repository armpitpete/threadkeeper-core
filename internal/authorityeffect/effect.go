package authorityeffect

import "fmt"

type Effect string

const (
	None                 Effect = "none"
	EvidencePreservation Effect = "evidence_preservation"
	IntegrityAttestation Effect = "integrity_attestation"
	DerivedProjection    Effect = "derived_projection"
	AuthorityTransition  Effect = "authority_transition"
	AccessControl        Effect = "access_control"
)

type Declaration struct {
	Component string `json:"component"`
	Effect    Effect `json:"effect"`
}

func (d Declaration) Validate() error {
	if d.Component=="" { return fmt.Errorf("AUTHORITY_EFFECT_INVALID: component required") }
	switch d.Effect { case None,EvidencePreservation,IntegrityAttestation,DerivedProjection,AuthorityTransition,AccessControl: return nil; default: return fmt.Errorf("AUTHORITY_EFFECT_INVALID: unknown effect %q",d.Effect) }
}
