package reviewbundle

import (
	"fmt"
	"sort"

	"github.com/armpitpete/threadkeeper-core/internal/authorityeffect"
	"github.com/armpitpete/threadkeeper-core/internal/decision"
	"github.com/armpitpete/threadkeeper-core/internal/evidence"
	"github.com/armpitpete/threadkeeper-core/internal/proposal"
	"github.com/armpitpete/threadkeeper-core/internal/simulation"
)

type Bundle struct {
	Proposal        proposal.Proposal `json:"proposal"`
	Evidence        []evidence.Envelope `json:"evidence"`
	DecisionContext decision.Context `json:"decision_context"`
	Simulation      simulation.Report `json:"simulation"`
	AuthorityEffect authorityeffect.Effect `json:"authority_effect"`
}

type Summary struct {
	ProposalID           string `json:"proposal_id"`
	TargetID             string `json:"target_id"`
	ExpectedStateVersion string `json:"expected_state_version"`
	Generated            bool   `json:"generated"`
	EvidenceCount        int    `json:"evidence_count"`
	ConflictCount        int    `json:"conflict_count"`
	AlternativesCount    int    `json:"alternatives_count"`
	DissentCount         int    `json:"dissent_count"`
	ReopeningCount       int    `json:"reopening_count"`
	ProjectionChanges    int    `json:"projection_changes"`
	AuthorityChanges     int    `json:"authority_changes"`
	NewConflicts         int    `json:"new_conflicts"`
	AccessChanges        int    `json:"access_changes"`
	AuthorityEffect      string `json:"authority_effect"`
}

func (b Bundle) Validate() error {
	if err:=b.Proposal.Validate();err!=nil{return err}
	if b.AuthorityEffect!=authorityeffect.None{return fmt.Errorf("REVIEW_BUNDLE_INVALID: review bundle must have authority effect %q",authorityeffect.None)}
	if err:=b.DecisionContext.Validate();err!=nil{return err}
	ids:=make([]string,0,len(b.Evidence)); seen:=map[string]struct{}{}
	for _,env:=range b.Evidence{if err:=env.Validate();err!=nil{return err};if _,ok:=seen[env.RecordID];ok{return fmt.Errorf("REVIEW_BUNDLE_INVALID: duplicate evidence %q",env.RecordID)};seen[env.RecordID]=struct{}{};ids=append(ids,env.RecordID)}
	sort.Strings(ids)
	if len(ids)!=len(b.Proposal.EvidenceIDs){return fmt.Errorf("REVIEW_BUNDLE_INVALID: proposal/evidence set size mismatch")}
	for i,id:=range ids{if b.Proposal.EvidenceIDs[i]!=id{return fmt.Errorf("REVIEW_BUNDLE_INVALID: proposal evidence set does not match included evidence")}}
	return nil
}

func (b Bundle) Summary() (Summary,error) {
	if err:=b.Validate();err!=nil{return Summary{},err}
	conflicts:=0;for _,env:=range b.Evidence{conflicts+=len(env.Conflicts)}
	return Summary{ProposalID:b.Proposal.ID,TargetID:b.Proposal.TargetID,ExpectedStateVersion:b.Proposal.ExpectedStateVersion,Generated:b.Proposal.Generated,EvidenceCount:len(b.Evidence),ConflictCount:conflicts,AlternativesCount:len(b.DecisionContext.Alternatives),DissentCount:len(b.DecisionContext.Dissent),ReopeningCount:len(b.DecisionContext.ReopeningConditions),ProjectionChanges:len(b.Simulation.ProjectionChanged),AuthorityChanges:len(b.Simulation.AuthorityChanged),NewConflicts:len(b.Simulation.NewConflicts),AccessChanges:len(b.Simulation.AccessChanged),AuthorityEffect:string(b.AuthorityEffect)},nil
}
