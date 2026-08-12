package selfcheck

import (
	"os"
	"strings"
	"testing"
)

func TestImplementationStatusMatchesExecutableSafetyBoundary(t *testing.T){
	raw,err:=os.ReadFile("../../IMPLEMENTATION_STATUS.md");if err!=nil{t.Fatal(err)}
	text:=string(raw)
	for _,required:=range []string{"Git ledger replay","candidate-write","AUTHORITY_WRITES_DISABLED","genesis","coverage","recovery-fork","destructive non-empty bare-ledger","source adapter","provenance graph","portable Core export","reference CLI","overload signalling"}{if !strings.Contains(text,required){t.Fatalf("implementation status missing %q",required)}}
	for _,stale:=range []string{"durable Git ledger open/fsck;\n- event replay/projections;","source-adapter escrow backends and broad temporal/coverage schema migration are not yet installed","federation transport and executable reference client are not yet built"}{if strings.Contains(text,stale){t.Fatalf("implementation status contains stale claim %q",stale)}}
}
