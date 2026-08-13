package escrow

import (
	"path/filepath"
	"testing"
)

func TestContentAddressedEscrowStore(t *testing.T){s,err:=OpenStore(filepath.Join(t.TempDir(),"escrow"));if err!=nil{t.Fatal(err)};snap,err:=s.Put("source:a","v1","text/plain",[]byte("evidence"));if err!=nil{t.Fatal(err)};got,err:=s.Get(snap);if err!=nil{t.Fatal(err)};if string(got)!="evidence"{t.Fatalf("got %q",got)};snap2,err:=s.Put("source:b","v2","text/plain",[]byte("evidence"));if err!=nil{t.Fatal(err)};if snap2.ContentSHA256!=snap.ContentSHA256{t.Fatal("same bytes should share content identity")}}
