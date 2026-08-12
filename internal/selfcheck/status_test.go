package selfcheck

import (
	"os"
	"strings"
	"testing"
)

func TestImplementationStatusMatchesExecutableSafetyBoundary(t *testing.T){
	raw,err:=os.ReadFile("../../IMPLEMENTATION_STATUS.md"); if err!=nil{t.Fatal(err)}
	text:=string(raw)
	for _,required:=range []string{"Git ledger replay","candidate-write","AUTHORITY_WRITES_DISABLED","genesis","coverage","fork recovery"}{ if !strings.Contains(text,required){t.Fatalf("implementation status missing %q",required)} }
	if strings.Contains(text,"durable Git ledger open/fsck;\n- event replay/projections;"){t.Fatal("implementation status still claims installed ledger/replay are absent")}
}
