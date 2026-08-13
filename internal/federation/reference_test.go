package federation

import "testing"

func TestFederationRequiresLocalAuthorityDisposition(t *testing.T) {
	r:=Reference{ProjectID:"p",LedgerID:"l",RecordID:"r",VersionID:"v",SourceAuthorityClass:"authoritative"}
	if err:=r.Validate(); err==nil{t.Fatal("expected missing local authority rejection")}
	r.LocalAuthorityClass="derived"; if err:=r.Validate(); err!=nil{t.Fatal(err)}
}
