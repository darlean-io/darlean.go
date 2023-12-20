package main

import (
	"core/backoff"
	"core/invoke"
	"core/natstransport"
	"core/portal"
	"core/remoteactorregistry"
	"core/transporthandler"
	"core/variant"
	"fmt"
	"time"
)

type EchoActor_Echo struct {
	A0     string
	A1     int
	A2     any
	Result string
}

type EchoActor_Greet struct {
	A0     string
	A1     int
	A2     any
	Result string
}

type EchoActor struct {
	Echo  EchoActor_Echo
	Greet EchoActor_Greet
}

func toLowerCase(invoker *invoke.DynamicInvoker, input string) {
	req := invoke.InvokeRequest{
		ActorType:  "echoactor",
		ActorId:    []string{"A"},
		ActionName: "echo",
		Parameters: []any{input},
	}

	value, err := invoker.Invoke(&req)
	if err != nil {
		fmt.Printf("Error for %v: %v\n", input, err)
		panic(err)
	}
	var v string
	err2 := variant.Assign(value, &v)
	if err2 != nil {
		panic(err)
	}
	fmt.Printf("Received: %v -> %v\n", input, v)
}

func main() {

	const OUR_APP_ID = "client"
	const NATS_ADDR = "localhost:4500"
	HOSTS := []string{"server01"}

	transport, err := natstransport.New(NATS_ADDR, OUR_APP_ID)
	if err != nil {
		panic(err)
	}

	staticInvoker := transporthandler.New(transport, nil, OUR_APP_ID)
	registry := remoteactorregistry.New(HOSTS, staticInvoker)

	backoff := backoff.Exponential(10*time.Millisecond, 8, 4.0, 0.25)
	invoker := invoke.NewDynamicInvoker(staticInvoker, backoff, registry)

	time.Sleep(time.Second)

	p := portal.New(&invoker)
	echoPortal := portal.ForType[EchoActor](p)
	actor := echoPortal.Obtain([]string{"abc"})
	call := actor.Call().Echo
	call.A0 = "Hello"
	call.A2 = 42
	err = actor.Invoke(&call)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Received via Portal: %v\n", call.Result)

	time.Sleep(time.Second)
	go toLowerCase(&invoker, "Hello")
	go toLowerCase(&invoker, "World")

	time.Sleep(15 * time.Second)
	registry.Stop()
	transport.Stop()
}
