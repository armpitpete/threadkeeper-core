package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/armpitpete/threadkeeper-core/internal/canonicaljson"
	"github.com/armpitpete/threadkeeper-core/internal/digest"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/ledger"
	"github.com/armpitpete/threadkeeper-core/internal/schema"
	"github.com/armpitpete/threadkeeper-core/internal/service"
	"github.com/armpitpete/threadkeeper-core/internal/strictjson"
)

var (
	version      = "dev"
	sourceCommit = "unknown"
)

type buildInfo struct {
	Version                string            `json:"version"`
	SourceCommit           string            `json:"source_commit"`
	GoVersion              string            `json:"go_version"`
	Platform               string            `json:"platform"`
	AuthorityWritesEnabled bool              `json:"authority_writes_enabled"`
	Dependencies           map[string]string `json:"dependencies"`
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "version":
		err = printVersion()
	case "check-json":
		err = oneFile(os.Args[2:], func(b []byte) error { return strictjson.Validate(b) })
	case "canonicalize":
		err = oneFileOut(os.Args[2:], canonicaljson.Canonicalize)
	case "digest":
		err = digestCommand(os.Args[2:])
	case "validate":
		err = validateCommand(os.Args[2:])
	case "ledger-inspect":
		err = ledgerInspectCommand(os.Args[2:])
	case "ledger-recovery-proof":
		err = ledgerRecoveryProofCommand(os.Args[2:])
	case "authority-write":
		err = service.RequireAuthorityWritesEnabled()
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: threadkeeper-core <version|check-json|canonicalize|digest|validate|ledger-inspect|ledger-recovery-proof|authority-write> ...")
}

func printVersion() error {
	deps := map[string]string{}
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range bi.Deps {
			deps[dep.Path] = dep.Version
		}
	}
	return json.NewEncoder(os.Stdout).Encode(buildInfo{
		Version: version, SourceCommit: sourceCommit, GoVersion: runtime.Version(),
		Platform: runtime.GOOS + "/" + runtime.GOARCH,
		AuthorityWritesEnabled: service.AuthorityWritesEnabled(), Dependencies: deps,
	})
}

func oneFile(args []string, fn func([]byte) error) error {
	if len(args) != 1 { return fmt.Errorf("expected one file argument") }
	b, err := os.ReadFile(args[0]); if err != nil { return err }
	return fn(b)
}

func oneFileOut(args []string, fn func([]byte) ([]byte, error)) error {
	if len(args) != 1 { return fmt.Errorf("expected one file argument") }
	b, err := os.ReadFile(args[0]); if err != nil { return err }
	out, err := fn(b); if err != nil { return err }
	_, err = os.Stdout.Write(append(out, '\n')); return err
}

func digestCommand(args []string) error {
	if len(args) != 1 { return fmt.Errorf("expected one file argument") }
	b, err := os.ReadFile(args[0]); if err != nil { return err }
	out, _, err := digest.Complete(b); if err != nil { return err }
	_, err = os.Stdout.Write(append(out, '\n')); return err
}

func validateCommand(args []string) error {
	if len(args) != 2 { return fmt.Errorf("expected schema file and instance file") }
	schemaBytes, err := os.ReadFile(args[0]); if err != nil { return err }
	instanceBytes, err := os.ReadFile(args[1]); if err != nil { return err }
	v, err := strictjson.Decode(schemaBytes); if err != nil { return err }
	obj, ok := v.(map[string]any); if !ok { return fmt.Errorf("schema root must be object") }
	id, _ := obj["$id"].(string); if id == "" { return fmt.Errorf("schema $id is required") }
	r := schema.NewRegistry(); if err := r.Add(id, schemaBytes); err != nil { return err }
	return r.Validate(id, instanceBytes)
}

func ledgerInspectCommand(args []string) error {
	r, err := ledgerReaderFromArgs(args)
	if err != nil { return err }
	manifest, err := ledger.Replay(context.Background(), r)
	if err != nil { return err }
	return writeIndentedJSON(manifest)
}

func ledgerRecoveryProofCommand(args []string) error {
	r, err := ledgerReaderFromArgs(args)
	if err != nil { return err }
	proof, err := ledger.ProveRecovery(context.Background(), r)
	if err != nil { return err }
	return writeIndentedJSON(proof)
}

func ledgerReaderFromArgs(args []string) (*gitledger.Reader, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("expected ledger git directory and optional authoritative ref")
	}
	ref := gitledger.DefaultRef
	if len(args) == 2 { ref = args[1] }
	return gitledger.New(args[0], ref)
}

func writeIndentedJSON(value any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}
