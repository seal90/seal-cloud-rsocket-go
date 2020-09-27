package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx"
	"github.com/rsocket/rsocket-go/rx/flux"
)

func TestFlux(t *testing.T) {
	p1 := payload.NewString("1", "metadata string")
	p2 := payload.NewString("2", "metadata string")
	f := flux.Just(p1, p2)
	f = f.SwitchOnFirst(func(s flux.Signal, f flux.Flux) flux.Flux {
		t := s.Type().String()
		p, b := s.Value()
		fmt.Println("t:", t, ", p : ", p.DataUTF8(), ", b :", b)
		return f
	})
	ctx := context.Background()
	f = f.Map(func(p payload.Payload) (payload.Payload, error) {
		fmt.Println("map:", p.DataUTF8())
		return p, nil
	})
	f.Subscribe(ctx, rx.OnNext(func(input payload.Payload) error {
		fmt.Println("sub:", input.DataUTF8())
		return nil
	}))
	// f.BlockLast(ctx)
	time.Sleep(10 * time.Second)
}
