package sourceadapter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/armpitpete/threadkeeper-core/internal/escrow"
	"github.com/armpitpete/threadkeeper-core/internal/source"
)

type FileAdapter struct{root string}

type Request struct {
	SourceID       string `json:"source_id"`
	RelativePath   string `json:"relative_path"`
	ExpectedSHA256 string `json:"expected_sha256"`
	MediaType      string `json:"media_type,omitempty"`
}

type Snapshot struct {
	Version source.Version `json:"version"`
	Content []byte         `json:"-"`
	MediaType string       `json:"media_type,omitempty"`
}

func NewFileAdapter(root string)(*FileAdapter,error){
	if root==""{return nil,fmt.Errorf("SOURCE_ADAPTER_INVALID: root required")}
	abs,err:=filepath.Abs(root);if err!=nil{return nil,err}
	info,err:=os.Lstat(abs);if err!=nil{return nil,err};if info.Mode()&os.ModeSymlink!=0||!info.IsDir(){return nil,fmt.Errorf("SOURCE_ADAPTER_INVALID: root must be a real directory")}
	return &FileAdapter{root:abs},nil
}

func(a *FileAdapter)Fetch(ctx context.Context,req Request)(Snapshot,error){
	if err:=ctx.Err();err!=nil{return Snapshot{},err}
	if req.SourceID==""||req.ExpectedSHA256==""{return Snapshot{},fmt.Errorf("SOURCE_ADAPTER_INVALID: source id and exact digest required")}
	clean:=filepath.Clean(filepath.FromSlash(req.RelativePath));if req.RelativePath==""||filepath.IsAbs(clean)||clean=="."||clean==".."||strings.HasPrefix(clean,".."+string(filepath.Separator)){return Snapshot{},fmt.Errorf("SOURCE_ADAPTER_INVALID: unsafe relative path")}
	full:=filepath.Join(a.root,clean);if err:=rejectSymlinkPath(a.root,clean);err!=nil{return Snapshot{},err}
	info,err:=os.Lstat(full);if err!=nil{return Snapshot{},err};if info.Mode()&os.ModeSymlink!=0||!info.Mode().IsRegular(){return Snapshot{},fmt.Errorf("SOURCE_ADAPTER_INVALID: source path must be a regular file")}
	content,err:=os.ReadFile(full);if err!=nil{return Snapshot{},err};got:=escrow.HashContent(content);if got!=strings.ToLower(req.ExpectedSHA256){return Snapshot{},fmt.Errorf("SOURCE_VERSION_MISMATCH: got %s want %s",got,req.ExpectedSHA256)}
	return Snapshot{Version:source.Version{ID:got,ContentSHA256:got,Locator:filepath.ToSlash(clean)},Content:content,MediaType:req.MediaType},nil
}

func FetchAndEscrow(ctx context.Context,adapter *FileAdapter,store *escrow.Store,req Request)(source.Version,escrow.Snapshot,error){
	snapshot,err:=adapter.Fetch(ctx,req);if err!=nil{return source.Version{},escrow.Snapshot{},err};if store==nil{return source.Version{},escrow.Snapshot{},fmt.Errorf("SOURCE_ADAPTER_INVALID: escrow store required")};stored,err:=store.Put(req.SourceID,snapshot.Version.ID,req.MediaType,snapshot.Content);if err!=nil{return source.Version{},escrow.Snapshot{},err};return snapshot.Version,stored,nil
}

func rejectSymlinkPath(root,rel string)error{current:=root;for _,part:=range strings.Split(filepath.Clean(rel),string(filepath.Separator)){current=filepath.Join(current,part);info,err:=os.Lstat(current);if err!=nil{return err};if info.Mode()&os.ModeSymlink!=0{return fmt.Errorf("SOURCE_ADAPTER_INVALID: symlink path component %q",part)}};return nil}
