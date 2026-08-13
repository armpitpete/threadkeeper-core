package proposal

import (
	"fmt"
	"sort"
	"time"

	"github.com/armpitpete/threadkeeper-core/internal/strictjson"
)

type Proposal struct {
	ID                   string   `json:"id"`
	ProposedBy           string   `json:"proposed_by"`
	Generated            bool     `json:"generated"`
	TargetID             string   `json:"target_id"`
	ExpectedStateVersion string   `json:"expected_state_version"`
	EvidenceIDs          []string `json:"evidence_ids,omitempty"`
	Rationale            string   `json:"rationale"`
	CreatedAt            string   `json:"created_at"`
	IdempotencyKey       string   `json:"idempotency_key"`
	Content              []byte   `json:"content"`
}

func (p Proposal) Validate() error {
	if p.ID=="" || p.ProposedBy=="" || p.TargetID=="" || p.ExpectedStateVersion=="" || p.Rationale=="" || p.IdempotencyKey=="" { return fmt.Errorf("PROPOSAL_INVALID: identity, actor, target, expected state, rationale and idempotency key are required") }
	if _,err:=time.Parse(time.RFC3339,p.CreatedAt);err!=nil{return fmt.Errorf("PROPOSAL_INVALID: created_at: %w",err)}
	if len(p.Content)==0{return fmt.Errorf("PROPOSAL_INVALID: content is required")}
	if err:=strictjson.Validate(p.Content);err!=nil{return fmt.Errorf("PROPOSAL_INVALID: content: %w",err)}
	if !sort.StringsAreSorted(p.EvidenceIDs){return fmt.Errorf("PROPOSAL_INVALID: evidence_ids must be sorted")}
	for i,id:=range p.EvidenceIDs{if id==""||(i>0&&p.EvidenceIDs[i-1]==id){return fmt.Errorf("PROPOSAL_INVALID: evidence_ids must be distinct and non-empty")}}
	return nil
}
