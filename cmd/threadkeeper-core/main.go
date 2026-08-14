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
	"github.com/armpitpete/threadkeeper-core/internal/evidence"
	"github.com/armpitpete/threadkeeper-core/internal/genesis"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/health"
	"github.com/armpitpete/threadkeeper-core/internal/ledger"
	"github.com/armpitpete/threadkeeper-core/internal/loadproof"
	"github.com/armpitpete/threadkeeper-core/internal/portable"
	"github.com/armpitpete/threadkeeper-core/internal/reviewbundle"
	"github.com/armpitpete/threadkeeper-core/internal/schema"
	"github.com/armpitpete/threadkeeper-core/internal/service"
	"github.com/armpitpete/threadkeeper-core/internal/sourceadapter"
	"github.com/armpitpete/threadkeeper-core/internal/strictjson"
)

var(version="dev";sourceCommit="unknown")
type buildInfo struct{Version string `json:"version"`;SourceCommit string `json:"source_commit"`;GoVersion string `json:"go_version"`;Platform string `json:"platform"`;AuthorityWritesEnabled bool `json:"authority_writes_enabled"`;Dependencies map[string]string `json:"dependencies"`}

func main(){if len(os.Args)<2{usage();os.Exit(2)};var err error;switch os.Args[1]{case"version":err=printVersion();case"check-json":err=oneFile(os.Args[2:],func(b[]byte)error{return strictjson.Validate(b)});case"canonicalize":err=oneFileOut(os.Args[2:],canonicaljson.Canonicalize);case"digest":err=digestCommand(os.Args[2:]);case"validate":err=validateCommand(os.Args[2:]);case"genesis-check":err=genesisCheckCommand(os.Args[2:]);case"fresh-genesis-init":err=freshGenesisInitCommand(os.Args[2:]);case"evidence-check":err=evidenceCheckCommand(os.Args[2:]);case"review-bundle":err=reviewBundleCommand(os.Args[2:]);case"health-check":err=healthCheckCommand(os.Args[2:]);case"source-file-check":err=sourceFileCheckCommand(os.Args[2:]);case"portable-check":err=portableCheckCommand(os.Args[2:]);case"ledger-inspect":err=ledgerInspectCommand(os.Args[2:]);case"ledger-recovery-proof":err=ledgerRecoveryProofCommand(os.Args[2:]);case"ledger-load-proof":err=ledgerLoadProofCommand(os.Args[2:]);case"recovery-compare":err=recoveryCompareCommand(os.Args[2:]);case"authority-write":err=service.RequireAuthorityWritesEnabled();default:usage();os.Exit(2)};if err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(1)}}
func usage(){fmt.Fprintln(os.Stderr,"usage: threadkeeper-core <version|check-json|canonicalize|digest|validate|genesis-check|fresh-genesis-init|evidence-check|review-bundle|health-check|source-file-check|portable-check|ledger-inspect|ledger-recovery-proof|ledger-load-proof|recovery-compare|authority-write> ...")}
func printVersion()error{deps:=map[string]string{};if bi,ok:=debug.ReadBuildInfo();ok{for _,dep:=range bi.Deps{deps[dep.Path]=dep.Version}};return json.NewEncoder(os.Stdout).Encode(buildInfo{Version:version,SourceCommit:sourceCommit,GoVersion:runtime.Version(),Platform:runtime.GOOS+"/"+runtime.GOARCH,AuthorityWritesEnabled:service.AuthorityWritesEnabled(),Dependencies:deps})}
func oneFile(args[]string,fn func([]byte)error)error{if len(args)!=1{return fmt.Errorf("expected one file argument")};b,err:=os.ReadFile(args[0]);if err!=nil{return err};return fn(b)}
func oneFileOut(args[]string,fn func([]byte)([]byte,error))error{if len(args)!=1{return fmt.Errorf("expected one file argument")};b,err:=os.ReadFile(args[0]);if err!=nil{return err};out,err:=fn(b);if err!=nil{return err};_,err=os.Stdout.Write(append(out,'\n'));return err}
func digestCommand(args[]string)error{if len(args)!=1{return fmt.Errorf("expected one file argument")};b,err:=os.ReadFile(args[0]);if err!=nil{return err};out,_,err:=digest.Complete(b);if err!=nil{return err};_,err=os.Stdout.Write(append(out,'\n'));return err}
func validateCommand(args[]string)error{if len(args)!=2{return fmt.Errorf("expected schema file and instance file")};schemaBytes,err:=os.ReadFile(args[0]);if err!=nil{return err};instanceBytes,err:=os.ReadFile(args[1]);if err!=nil{return err};v,err:=strictjson.Decode(schemaBytes);if err!=nil{return err};obj,ok:=v.(map[string]any);if !ok{return fmt.Errorf("schema root must be object")};id,_:=obj["$id"].(string);if id==""{return fmt.Errorf("schema $id is required")};r:=schema.NewRegistry();if err:=r.Add(id,schemaBytes);err!=nil{return err};return r.Validate(id,instanceBytes)}
func genesisCheckCommand(args[]string)error{if len(args)!=1{return fmt.Errorf("expected genesis file")};raw,err:=os.ReadFile(args[0]);if err!=nil{return err};root,err:=genesis.Validate(raw);if err!=nil{return err};return writeIndentedJSON(root)}
func evidenceCheckCommand(args[]string)error{var env evidence.Envelope;if err:=decodeStrictFile(args,&env);err!=nil{return err};if err:=env.Validate();err!=nil{return err};return writeIndentedJSON(env)}
func reviewBundleCommand(args[]string)error{var bundle reviewbundle.Bundle;if err:=decodeStrictFile(args,&bundle);err!=nil{return err};summary,err:=bundle.Summary();if err!=nil{return err};return writeIndentedJSON(summary)}
func healthCheckCommand(args[]string)error{var checks[]health.Check;if err:=decodeStrictFile(args,&checks);err!=nil{return err};status,err:=health.Aggregate(checks);if err!=nil{return err};return writeIndentedJSON(map[string]any{"status":status,"checks":checks})}
func sourceFileCheckCommand(args[]string)error{if len(args)!=3{return fmt.Errorf("expected source root, relative path and exact SHA-256")};adapter,err:=sourceadapter.NewFileAdapter(args[0]);if err!=nil{return err};snapshot,err:=adapter.Fetch(context.Background(),sourceadapter.Request{SourceID:"cli:source",RelativePath:args[1],ExpectedSHA256:args[2]});if err!=nil{return err};return writeIndentedJSON(snapshot.Version)}
func portableCheckCommand(args[]string)error{if len(args)!=1{return fmt.Errorf("expected portable export file")};raw,err:=os.ReadFile(args[0]);if err!=nil{return err};bundle,err:=portable.Decode(raw);if err!=nil{return err};return writeIndentedJSON(map[string]any{"format":bundle.Format,"ledger_commit":bundle.LedgerCommit,"sources":len(bundle.Sources),"provenance_records":len(bundle.Provenance),"relationships":len(bundle.Relationships),"conflicts":len(bundle.Conflicts),"evidence_records":len(bundle.Evidence)})}
func recoveryCompareCommand(args[]string)error{if len(args)!=2{return fmt.Errorf("expected original and restored recovery-proof files")};var a,b ledger.RecoveryProof;if err:=decodeStrictPath(args[0],&a);err!=nil{return err};if err:=decodeStrictPath(args[1],&b);err!=nil{return err};if err:=ledger.CompareRecoveryProofs(a,b);err!=nil{return err};return writeIndentedJSON(map[string]bool{"equivalent":true})}
func ledgerInspectCommand(args[]string)error{r,err:=ledgerReaderFromArgs(args);if err!=nil{return err};defer r.Close();manifest,err:=ledger.Replay(context.Background(),r);if err!=nil{return err};return writeIndentedJSON(manifest)}
func ledgerRecoveryProofCommand(args[]string)error{r,err:=ledgerReaderFromArgs(args);if err!=nil{return err};defer r.Close();proof,err:=ledger.ProveRecovery(context.Background(),r);if err!=nil{return err};return writeIndentedJSON(proof)}
func ledgerLoadProofCommand(args[]string)error{if len(args)<2||len(args)>3{return fmt.Errorf("expected ledger git directory, envelope file and optional authoritative ref")};raw,err:=os.ReadFile(args[1]);if err!=nil{return err};envelope,err:=loadproof.DecodeEnvelope(raw);if err!=nil{return err};readerArgs:=[]string{args[0]};if len(args)==3{readerArgs=append(readerArgs,args[2])};r,err:=ledgerReaderFromArgs(readerArgs);if err!=nil{return err};defer r.Close();proof,proofErr:=ledger.ProveReadLoad(context.Background(),r,envelope);if proof!=nil{if err:=writeIndentedJSON(proof);err!=nil{return err}};return proofErr}
func ledgerReaderFromArgs(args[]string)(*gitledger.Reader,error){if len(args)<1||len(args)>2{return nil,fmt.Errorf("expected ledger git directory and optional authoritative ref")};ref:=gitledger.DefaultRef;if len(args)==2{ref=args[1]};return gitledger.New(args[0],ref)}
func decodeStrictFile(args[]string,value any)error{if len(args)!=1{return fmt.Errorf("expected one file argument")};return decodeStrictPath(args[0],value)}
func decodeStrictPath(path string,value any)error{raw,err:=os.ReadFile(path);if err!=nil{return err};if err:=strictjson.Validate(raw);err!=nil{return err};if err:=json.Unmarshal(raw,value);err!=nil{return err};return nil}
func writeIndentedJSON(value any)error{enc:=json.NewEncoder(os.Stdout);enc.SetIndent("","  ");return enc.Encode(value)}
