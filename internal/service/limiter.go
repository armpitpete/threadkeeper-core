package service

import (
	"errors"
	"fmt"
)

var ErrOverloaded = errors.New("OVERLOADED")

type Limiter struct{ slots chan struct{} }

func NewLimiter(capacity int)(*Limiter,error){if capacity<=0{return nil,fmt.Errorf("invalid limiter capacity %d",capacity)};return &Limiter{slots:make(chan struct{},capacity)},nil}

// TryAcquire fails explicitly rather than queueing without bound. The returned
// release function must be called exactly once by a successful caller.
func(l *Limiter)TryAcquire()(func(),error){select{case l.slots<-struct{}{}:released:=false;return func(){if !released{released=true;<-l.slots}},nil;default:return nil,ErrOverloaded}}
