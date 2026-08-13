package reviewbundle

import (
	"testing"
	"github.com/armpitpete/threadkeeper-core/internal/access"
	"github.com/armpitpete/threadkeeper-core/internal/authorityeffect"
	"github.com/armpitpete/threadkeeper-core/internal/evidence"
	"github.com/armpitpete/threadkeeper-core/internal/proposal"
)

func TestReviewBundleCannotBecomeAuthorityByInspection(t *testing.T){b:=Bundle{Proposal:proposal.Proposal{ID:"p",ProposedBy:"human:a",TargetID:"x",ExpectedStateVersion:"h",EvidenceIDs:[]string{"e"},Rationale:"r",CreatedAt:"2026-08-12T17:00:00Z",IdempotencyKey:"i",Content:[]byte(`{"x":1}`)},Evidence:[]evidence.Envelope{{RecordID:"e",RecordType:"source",AuthorityClass:"authoritative",Classification:access.Public}},AuthorityEffect:authorityeffect.None};s,err:=b.Summary();if err!=nil{t.Fatal(err)};if s.AuthorityEffect!="none"{t.Fatalf("unexpected effect %#v",s)};b.AuthorityEffect=authorityeffect.AuthorityTransition;if _,err:=b.Summary();err==nil{t.Fatal("review must not be an authority transition")}}
