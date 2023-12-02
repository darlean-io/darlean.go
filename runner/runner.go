package main

import (
	"core/backoff"
	"core/invoke"
	"core/natstransport"
	"core/remoteactorregistry"
	"core/variant"
	"fmt"
	"time"
)

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

	staticInvoker := invoke.NewStaticInvoker(transport, OUR_APP_ID)
	registry := remoteactorregistry.New(HOSTS, staticInvoker)

	backoff := backoff.Exponential(10*time.Millisecond, 8, 4.0, 0.25)
	invoker := invoke.NewDynamicInvoker(staticInvoker, backoff, registry)

	time.Sleep(time.Second)
	go toLowerCase(&invoker, "Hello")
	go toLowerCase(&invoker, "World")

	time.Sleep(15 * time.Second)
	registry.Stop()
	transport.Stop()
}
