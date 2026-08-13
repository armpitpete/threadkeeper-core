package proposal

import "testing"

func TestProposalIsReviewableButNotAuthority(t *testing.T){p:=Proposal{ID:"p1",ProposedBy:"ai:reviewer",Generated:true,TargetID:"setting:x",ExpectedStateVersion:"head1",EvidenceIDs:[]string{"e1"},Rationale:"test candidate",CreatedAt:"2026-08-12T17:00:00Z",IdempotencyKey:"idem1",Content:[]byte(`{"value":1}`)};if err:=p.Validate();err!=nil{t.Fatal(err)}}
