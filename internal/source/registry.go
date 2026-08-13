package source

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/armpitpete/threadkeeper-core/internal/access"
	"github.com/armpitpete/threadkeeper-core/internal/escrow"
)

type Version struct {
	ID            string `json:"id"`
	ContentSHA256 string `json:"content_sha256,omitempty"`
	Locator       string `json:"locator,omitempty"`
	ObservedAt    string `json:"observed_at,omitempty"`
}

type Source struct {
	ID             string                `json:"id"`
	Kind           string                `json:"kind"`
	AuthorityClass string                `json:"authority_class"`
	Classification access.Classification `json:"classification"`
	Preservation   escrow.Policy         `json:"preservation"`
	Versions       map[string]Version    `json:"versions,omitempty"`
}

type Registry struct { sources map[string]Source }

func NewRegistry() *Registry { return &Registry{sources: map[string]Source{}} }

func (r *Registry) Register(src Source) error {
	if src.ID=="" || src.Kind=="" || src.AuthorityClass=="" { return fmt.Errorf("SOURCE_INVALID: id, kind and authority_class required") }
	if err:=src.Classification.Validate(); err!=nil{return err}
	if err:=src.Preservation.Validate(); err!=nil{return err}
	if _,exists:=r.sources[src.ID]; exists { return fmt.Errorf("SOURCE_EXISTS: %s",src.ID) }
	if src.Versions==nil { src.Versions=map[string]Version{} }
	for id,v:=range src.Versions { if id!=v.ID || v.ID=="" { return fmt.Errorf("SOURCE_VERSION_INVALID: map key/version id mismatch") } }
	r.sources[src.ID]=clone(src)
	return nil
}

func (r *Registry) AddVersion(sourceID string, version Version) error {
	src,ok:=r.sources[sourceID]; if !ok{return fmt.Errorf("SOURCE_NOT_FOUND: %s",sourceID)}
	if version.ID=="" {return fmt.Errorf("SOURCE_VERSION_INVALID: id required")}
	if prior,exists:=src.Versions[version.ID]; exists { if reflect.DeepEqual(prior,version){return nil}; return fmt.Errorf("SOURCE_VERSION_CONFLICT: %s@%s",sourceID,version.ID) }
	src.Versions[version.ID]=version; r.sources[sourceID]=src; return nil
}

func (r *Registry) Get(sourceID string) (Source,bool) { src,ok:=r.sources[sourceID]; return clone(src),ok }

func (r *Registry) ExactVersion(sourceID,versionID string) (Version,bool) { src,ok:=r.sources[sourceID]; if !ok{return Version{},false}; v,ok:=src.Versions[versionID]; return v,ok }

func (r *Registry) IDs() []string { out:=make([]string,0,len(r.sources)); for id:=range r.sources{out=append(out,id)}; sort.Strings(out); return out }

func clone(src Source) Source { out:=src; out.Versions=map[string]Version{}; for k,v:=range src.Versions{out.Versions[k]=v}; return out }
