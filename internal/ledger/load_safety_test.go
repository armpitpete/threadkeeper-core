package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
)

func TestConcurrentReplayReturnsOneExactSnapshot(t *testing.T){
	r,_:=candidateTestReader(t)
	const workers=12
	proofs:=make(chan *RecoveryProof,workers);errs:=make(chan error,workers);var wg sync.WaitGroup
	for i:=0;i<workers;i++{wg.Add(1);go func(){defer wg.Done();p,err:=ProveRecovery(context.Background(),r);if err!=nil{errs<-err;return};proofs<-p}()}
	wg.Wait();close(proofs);close(errs);for err:=range errs{t.Fatal(err)}
	var first *RecoveryProof;count:=0;for p:=range proofs{count++;if first==nil{first=p;continue};if err:=CompareRecoveryProofs(*first,*p);err!=nil{t.Fatal(err)}};if count!=workers{t.Fatalf("got %d proofs want %d",count,workers)}
}

func TestConcurrentCompetingWritersAcceptAtMostOne(t *testing.T){
	r,head:=candidateTestReader(t)
	aEvent:=makeCreateCandidateEventForTarget(t,head,"load-a","load-idem-a","setting:a",json.RawMessage(`{"value":"a"}`))
	bEvent:=makeCreateCandidateEventForTarget(t,head,"load-b","load-idem-b","setting:b",json.RawMessage(`{"value":"b"}`))
	a,_,err:=PrepareWriteCandidate(context.Background(),r,CandidateRequest{ExpectedHead:head,EventPath:"events/governance/load-a.json",Event:aEvent});if err!=nil{t.Fatal(err)}
	b,_,err:=PrepareWriteCandidate(context.Background(),r,CandidateRequest{ExpectedHead:head,EventPath:"events/governance/load-b.json",Event:bEvent});if err!=nil{t.Fatal(err)}
	type result struct{response *WriteResponse;err error};results:=make(chan result,2);start:=make(chan struct{});var wg sync.WaitGroup
	for _,c:=range []WriteCandidate{*a,*b}{candidate:=c;wg.Add(1);go func(){defer wg.Done();<-start;resp,err:=AcceptWriteCandidate(context.Background(),r,candidate);results<-result{resp,err}}()}
	close(start);wg.Wait();close(results)
	accepted:=0;stale:=0;for got:=range results{if got.err==nil{if got.response==nil||got.response.Status!=WriteStatusAccepted{t.Fatalf("unexpected success %#v",got.response)};accepted++;continue};if errors.Is(got.err,gitledger.ErrStaleState){stale++;continue};t.Fatalf("unexpected competing-writer error %v",got.err)}
	if accepted!=1||stale!=1{t.Fatalf("accepted=%d stale=%d",accepted,stale)}
}

func TestConcurrentExactRetryIsOneAcceptance(t *testing.T){
	r,head:=candidateTestReader(t);event:=makeCreateCandidateEvent(t,head,"load-retry","load-idem-retry",json.RawMessage(`{"enabled":true}`));candidate,_,err:=PrepareWriteCandidate(context.Background(),r,CandidateRequest{ExpectedHead:head,EventPath:"events/governance/load-retry.json",Event:event});if err!=nil{t.Fatal(err)}
	type result struct{response *WriteResponse;err error};results:=make(chan result,2);start:=make(chan struct{});var wg sync.WaitGroup
	for i:=0;i<2;i++{wg.Add(1);go func(){defer wg.Done();<-start;resp,err:=AcceptWriteCandidate(context.Background(),r,*candidate);results<-result{resp,err}}()};close(start);wg.Wait();close(results)
	accepted:=0;retried:=0;for got:=range results{if got.err!=nil{t.Fatal(got.err)};switch got.response.Status{case WriteStatusAccepted:accepted++;case WriteStatusAlreadyAccepted:retried++;default:t.Fatalf("unexpected status %#v",got.response)}};if accepted!=1||retried!=1{t.Fatalf("accepted=%d retry=%d",accepted,retried)}
}

func TestConcurrentSameKeyConflictIsDeterministic(t *testing.T){
	r,head:=candidateTestReader(t);const key="load-idem-conflict"
	aEvent:=makeCreateCandidateEventForTarget(t,head,"load-conflict-a",key,"setting:a",json.RawMessage(`{"value":"a"}`));bEvent:=makeCreateCandidateEventForTarget(t,head,"load-conflict-b",key,"setting:b",json.RawMessage(`{"value":"b"}`))
	a,_,err:=PrepareWriteCandidate(context.Background(),r,CandidateRequest{ExpectedHead:head,EventPath:"events/governance/load-conflict-a.json",Event:aEvent});if err!=nil{t.Fatal(err)};b,_,err:=PrepareWriteCandidate(context.Background(),r,CandidateRequest{ExpectedHead:head,EventPath:"events/governance/load-conflict-b.json",Event:bEvent});if err!=nil{t.Fatal(err)}
	type result struct{response *WriteResponse;err error};results:=make(chan result,2);start:=make(chan struct{});var wg sync.WaitGroup
	for _,c:=range []WriteCandidate{*a,*b}{candidate:=c;wg.Add(1);go func(){defer wg.Done();<-start;resp,err:=AcceptWriteCandidate(context.Background(),r,candidate);results<-result{resp,err}}()};close(start);wg.Wait();close(results)
	accepted:=0;conflicts:=0;for got:=range results{if got.err==nil{if got.response.Status!=WriteStatusAccepted{t.Fatalf("unexpected success %#v",got.response)};accepted++;continue};if errors.Is(got.err,ErrIdempotencyConflict){conflicts++;continue};t.Fatalf("unexpected same-key race error %v",got.err)};if accepted!=1||conflicts!=1{t.Fatalf("accepted=%d conflicts=%d",accepted,conflicts)}
}

func TestCancelledPrepareCannotMoveAuthority(t *testing.T){
	r,head:=candidateTestReader(t);event:=makeCreateCandidateEvent(t,head,"load-cancel","load-idem-cancel",json.RawMessage(`{"enabled":true}`));ctx,cancel:=context.WithCancel(context.Background());cancel();_,_,err:=PrepareWriteCandidate(ctx,r,CandidateRequest{ExpectedHead:head,EventPath:"events/governance/load-cancel.json",Event:event});if err==nil{t.Fatal("expected cancelled prepare failure")};got,headErr:=r.Head(context.Background());if headErr!=nil{t.Fatal(headErr)};if got!=head{t.Fatalf("cancelled prepare moved authority: got %s want %s",got,head)}
}
