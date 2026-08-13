package service

import (
	"errors"
	"testing"
)

func TestLimiterReturnsExplicitOverload(t *testing.T){l,err:=NewLimiter(1);if err!=nil{t.Fatal(err)};release,err:=l.TryAcquire();if err!=nil{t.Fatal(err)};if _,err:=l.TryAcquire();!errors.Is(err,ErrOverloaded){t.Fatalf("expected overload, got %v",err)};release();if _,err:=l.TryAcquire();err!=nil{t.Fatalf("slot was not released: %v",err)}}
