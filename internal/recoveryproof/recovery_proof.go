package recoveryproof

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/armpitpete/threadkeeper-core/internal/ledger"
	"github.com/armpitpete/threadkeeper-core/internal/strictjson"
)

var requiredFields = []string{
	"ledger_commit",
	"authoritative_ref",
	"git_object_format",
	"genesis_commit",
	"project_id",
	"ledger_id",
	"genesis_content_sha256",
	"actor_policy_version",
	"actor_policy_root_content_sha256",
	"history_commit_count",
	"event_count",
	"reducer_binding_count",
	"governed_record_count",
	"governed_records_sha256",
	"replay_sha256",
}

// Decode parses one complete RecoveryProof document. It rejects missing, null,
// duplicate and unknown members before returning a proof value so callers never
// confuse JSON omissions with Go zero values.
func Decode(raw []byte) (ledger.RecoveryProof, error) {
	if err := strictjson.Validate(raw); err != nil {
		return ledger.RecoveryProof{}, fmt.Errorf("RECOVERY_PROOF_INVALID: %w", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return ledger.RecoveryProof{}, fmt.Errorf("RECOVERY_PROOF_INVALID: decode fields: %w", err)
	}
	for _, name := range requiredFields {
		value, ok := fields[name]
		if !ok {
			return ledger.RecoveryProof{}, fmt.Errorf("RECOVERY_PROOF_INVALID: required field %q is missing", name)
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return ledger.RecoveryProof{}, fmt.Errorf("RECOVERY_PROOF_INVALID: required field %q must not be null", name)
		}
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var proof ledger.RecoveryProof
	if err := decoder.Decode(&proof); err != nil {
		return ledger.RecoveryProof{}, fmt.Errorf("RECOVERY_PROOF_INVALID: decode: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return ledger.RecoveryProof{}, fmt.Errorf("RECOVERY_PROOF_INVALID: trailing JSON value")
		}
		return ledger.RecoveryProof{}, fmt.Errorf("RECOVERY_PROOF_INVALID: trailing data: %w", err)
	}
	if err := Validate(proof); err != nil {
		return ledger.RecoveryProof{}, err
	}
	return proof, nil
}

// Validate checks semantic completeness of a decoded RecoveryProof value.
func Validate(proof ledger.RecoveryProof) error {
	if proof.LedgerCommit == "" ||
		proof.AuthoritativeRef == "" ||
		proof.GitObjectFormat == "" ||
		proof.GenesisCommit == "" ||
		proof.ProjectID == "" ||
		proof.LedgerID == "" ||
		proof.GenesisContentSHA256 == "" ||
		proof.ActorPolicyVersion == "" ||
		proof.ActorPolicyRootContentSHA256 == "" ||
		proof.GovernedRecordsSHA256 == "" ||
		proof.ReplaySHA256 == "" {
		return fmt.Errorf("RECOVERY_PROOF_INVALID: proof lacks required authority identity")
	}
	if proof.HistoryCommitCount <= 0 ||
		proof.EventCount < 0 ||
		proof.ReducerBindingCount < 0 ||
		proof.GovernedRecordCount < 0 {
		return fmt.Errorf("RECOVERY_PROOF_INVALID: proof contains invalid negative/zero counts")
	}
	return nil
}
