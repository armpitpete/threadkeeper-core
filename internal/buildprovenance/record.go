package buildprovenance

import (
	"fmt"
	"strings"
)

type Record struct {
	Version        string            `json:"version"`
	SourceCommit   string            `json:"source_commit"`
	GoVersion      string            `json:"go_version"`
	Platform       string            `json:"platform"`
	BinarySHA256   string            `json:"binary_sha256"`
	SBOMSHA256     string            `json:"sbom_sha256,omitempty"`
	Dependencies   map[string]string `json:"dependencies,omitempty"`
}

func (r Record) Validate() error {
	if r.Version=="" || r.GoVersion=="" || r.Platform=="" { return fmt.Errorf("BUILD_PROVENANCE_INVALID: version, go_version and platform required") }
	if !isHexID(r.SourceCommit,40,64) { return fmt.Errorf("BUILD_PROVENANCE_INVALID: source_commit") }
	if !isHexID(r.BinarySHA256,64) { return fmt.Errorf("BUILD_PROVENANCE_INVALID: binary_sha256") }
	if r.SBOMSHA256!="" && !isHexID(r.SBOMSHA256,64) { return fmt.Errorf("BUILD_PROVENANCE_INVALID: sbom_sha256") }
	return nil
}

func isHexID(s string, lengths ...int) bool { ok:=false; for _,n:=range lengths{if len(s)==n{ok=true}}; if !ok{return false}; for _,c:=range strings.ToLower(s){if !((c>='0'&&c<='9')||(c>='a'&&c<='f')){return false}}; return true }
