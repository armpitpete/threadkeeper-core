package sourceadapter

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"github.com/armpitpete/threadkeeper-core/internal/escrow"
)

func TestExactFileAdapterAndEscrow(t *testing.T){root:=t.TempDir();if err:=os.WriteFile(filepath.Join(root,"evidence.txt"),[]byte("evidence"),0o600);err!=nil{t.Fatal(err)};a,err:=NewFileAdapter(root);if err!=nil{t.Fatal(err)};digest:=escrow.HashContent([]byte("evidence"));store,err:=escrow.OpenStore(filepath.Join(t.TempDir(),"escrow"));if err!=nil{t.Fatal(err)};version,snap,err:=FetchAndEscrow(context.Background(),a,store,Request{SourceID:"source:a",RelativePath:"evidence.txt",ExpectedSHA256:digest,MediaType:"text/plain"});if err!=nil{t.Fatal(err)};if version.ID!=digest||snap.ContentSHA256!=digest{t.Fatalf("unexpected identities %#v %#v",version,snap)}}

func TestFileAdapterRejectsChangedOrUnsafeSource(t *testing.T){root:=t.TempDir();if err:=os.WriteFile(filepath.Join(root,"x"),[]byte("new"),0o600);err!=nil{t.Fatal(err)};a,err:=NewFileAdapter(root);if err!=nil{t.Fatal(err)};if _,err:=a.Fetch(context.Background(),Request{SourceID:"s",RelativePath:"x",ExpectedSHA256:escrow.HashContent([]byte("old"))});err==nil{t.Fatal("expected exact-version mismatch")};if _,err:=a.Fetch(context.Background(),Request{SourceID:"s",RelativePath:"../x",ExpectedSHA256:"x"});err==nil{t.Fatal("expected traversal rejection")}}
