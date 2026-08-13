package buildprovenance

import "testing"

func TestBuildProvenanceValidation(t *testing.T){
	r:=Record{Version:"v1",SourceCommit:"0123456789012345678901234567890123456789",GoVersion:"go1.x",Platform:"linux/amd64",BinarySHA256:"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"}
	if err:=r.Validate(); err!=nil{t.Fatal(err)}
}
