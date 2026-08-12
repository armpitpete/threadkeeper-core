package evidence

import (
	"testing"
	"github.com/armpitpete/threadkeeper-core/internal/access"
)

func TestEvidenceAuthorityIndependentOfRetrievalScore(t *testing.T){ score:=0.99; e:=Envelope{RecordID:"r",RecordType:"claim",AuthorityClass:"derived",RetrievalScore:&score,Classification:access.Public}; if err:=e.Validate();err!=nil{t.Fatal(err)}; if e.AuthorityClass!="derived"{t.Fatal("retrieval score must not promote authority")}}
