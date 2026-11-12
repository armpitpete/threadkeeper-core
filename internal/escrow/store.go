package escrow

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Store struct{dir string}

func OpenStore(dir string)(*Store,error){if dir==""{return nil,fmt.Errorf("ESCROW_STORE_INVALID: directory required")};if info,err:=os.Lstat(dir);err==nil&&info.Mode()&os.ModeSymlink!=0{return nil,fmt.Errorf("ESCROW_STORE_INVALID: directory must not be symlink")};if err:=os.MkdirAll(dir,0o700);err!=nil{return nil,err};if err:=os.Chmod(dir,0o700);err!=nil{return nil,err};abs,err:=filepath.Abs(dir);if err!=nil{return nil,err};return &Store{dir:abs},nil}

func(s *Store)Put(sourceID,versionID,mediaType string,content []byte)(Snapshot,error){if sourceID==""||versionID==""{return Snapshot{},fmt.Errorf("ESCROW_INVALID: source/version identity required")};digest:=HashContent(content);path:=filepath.Join(s.dir,digest+".blob");f,err:=os.OpenFile(path,os.O_WRONLY|os.O_CREATE|os.O_EXCL,0o600);if errors.Is(err,os.ErrExist){existing,readErr:=s.readBlob(digest+".blob");if readErr!=nil{return Snapshot{},readErr};if HashContent(existing)!=digest{return Snapshot{},fmt.Errorf("ESCROW_INTEGRITY_FAILURE: existing blob does not match filename digest")};return Snapshot{SourceID:sourceID,VersionID:versionID,ContentSHA256:digest,Size:int64(len(content)),MediaType:mediaType},nil};if err!=nil{return Snapshot{},err};if _,err:=f.Write(content);err!=nil{f.Close();os.Remove(path);return Snapshot{},err};if err:=f.Close();err!=nil{os.Remove(path);return Snapshot{},err};return Snapshot{SourceID:sourceID,VersionID:versionID,ContentSHA256:digest,Size:int64(len(content)),MediaType:mediaType},nil}

func(s *Store)Get(snapshot Snapshot)([]byte,error){content,err:=s.readBlob(snapshot.ContentSHA256+".blob");if err!=nil{return nil,err};if err:=VerifyContent(snapshot,content);err!=nil{return nil,err};return content,nil}

// readBlob confines reads to the escrow directory even if another process
// replaces a digest-named file with a symlink between operations.
func (s *Store) readBlob(name string) ([]byte, error) {
	file, err := os.OpenInRoot(s.dir, name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("ESCROW_STORE_INVALID: blob must be a regular file")
	}
	return io.ReadAll(file)
}
