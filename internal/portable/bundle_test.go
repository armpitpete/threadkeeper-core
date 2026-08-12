package portable

import (
	"bytes"
	"testing"
	"github.com/armpitpete/threadkeeper-core/internal/access"
	"github.com/armpitpete/threadkeeper-core/internal/evidence"
	"github.com/armpitpete/threadkeeper-core/internal/ledger"
)

func TestPortableExportRoundTripsCanonically(t *testing.T){b:=Bundle{Format:FormatV1,LedgerCommit:"head",RecoveryProof:ledger.RecoveryProof{LedgerCommit:"head",AuthoritativeRef:"refs/heads/main",GitObjectFormat:"sha1",ReplaySHA256:"r",GovernedRecordsSHA256:"p"},Evidence:[]evidence.Envelope{{RecordID:"b",RecordType:"claim",AuthorityClass:"derived",Classification:access.Public},{RecordID:"a",RecordType:"source",AuthorityClass:"authoritative",Classification:access.Public}}};raw,err:=Encode(b);if err!=nil{t.Fatal(err)};decoded,err:=Decode(raw);if err!=nil{t.Fatal(err)};again,err:=Encode(decoded);if err!=nil{t.Fatal(err)};if !bytes.Equal(raw,again){t.Fatalf("portable bytes changed\n%s\n%s",raw,again)}}

func TestPortableImportRejectsNonCanonicalOrWrongHead(t *testing.T){raw:=[]byte(`{ "format":"urn:threadkeeper:portable-core:v1","ledger_commit":"x","recovery_proof":{"ledger_commit":"y"} }`);if _,err:=Decode(raw);err==nil{t.Fatal("expected noncanonical/identity failure")}}
