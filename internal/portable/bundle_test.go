package portable

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/access"
	"github.com/armpitpete/threadkeeper-core/internal/canonicaljson"
	"github.com/armpitpete/threadkeeper-core/internal/evidence"
	"github.com/armpitpete/threadkeeper-core/internal/ledger"
)

func TestPortableExportRoundTripsCanonically(t *testing.T){b:=Bundle{Format:FormatV1,LedgerCommit:"head",RecoveryProof:ledger.RecoveryProof{LedgerCommit:"head",AuthoritativeRef:"refs/heads/main",GitObjectFormat:"sha1",ReplaySHA256:"r",GovernedRecordsSHA256:"p"},Evidence:[]evidence.Envelope{{RecordID:"b",RecordType:"claim",AuthorityClass:"derived",Classification:access.Public},{RecordID:"a",RecordType:"source",AuthorityClass:"authoritative",Classification:access.Public}}};raw,err:=Encode(b);if err!=nil{t.Fatal(err)};decoded,err:=Decode(raw);if err!=nil{t.Fatal(err)};again,err:=Encode(decoded);if err!=nil{t.Fatal(err)};if !bytes.Equal(raw,again){t.Fatalf("portable bytes changed\n%s\n%s",raw,again)}}

func TestPortableImportRejectsNonCanonicalOrWrongHead(t *testing.T){raw:=[]byte(`{ "format":"urn:threadkeeper:portable-core:v1","ledger_commit":"x","recovery_proof":{"ledger_commit":"y"} }`);if _,err:=Decode(raw);err==nil{t.Fatal("expected noncanonical/identity failure")}}

func TestPortableImportRejectsUnknownFields(t *testing.T){
	b:=Bundle{Format:FormatV1,LedgerCommit:"head",RecoveryProof:ledger.RecoveryProof{LedgerCommit:"head",AuthoritativeRef:"refs/heads/main",GitObjectFormat:"sha1",ReplaySHA256:"r",GovernedRecordsSHA256:"p"}}
	raw,err:=Encode(b);if err!=nil{t.Fatal(err)}
	var obj map[string]any
	if err:=json.Unmarshal(raw,&obj);err!=nil{t.Fatal(err)}
	obj["unexpected"]=true
	withUnknown,err:=json.Marshal(obj);if err!=nil{t.Fatal(err)}
	withUnknown,err=canonicaljson.Canonicalize(withUnknown);if err!=nil{t.Fatal(err)}
	if _,err:=Decode(withUnknown);err==nil{t.Fatal("expected unknown-field rejection")}
}

func TestPortableImportRejectsCanonicalButUnnormalizedOrder(t *testing.T){
	b:=Bundle{Format:FormatV1,LedgerCommit:"head",RecoveryProof:ledger.RecoveryProof{LedgerCommit:"head",AuthoritativeRef:"refs/heads/main",GitObjectFormat:"sha1",ReplaySHA256:"r",GovernedRecordsSHA256:"p"},Evidence:[]evidence.Envelope{{RecordID:"b",RecordType:"claim",AuthorityClass:"derived",Classification:access.Public},{RecordID:"a",RecordType:"source",AuthorityClass:"authoritative",Classification:access.Public}}}
	raw,err:=json.Marshal(b);if err!=nil{t.Fatal(err)}
	raw,err=canonicaljson.Canonicalize(raw);if err!=nil{t.Fatal(err)}
	if _,err:=Decode(raw);err==nil{t.Fatal("expected deterministic-order rejection")}
}
