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
		Parameters: []variant.Variant{variant.New(input)},
	}

	response := invoker.Invoke(&req)
	if response.Error != nil {
		var value any
		value, err := response.Error.Get(value)
		if err != nil {
			panic(err)
		}
		if value != nil {
			fmt.Printf("Error for %v: %v", input, value)
			panic(err)
		}
	}
	var value any
	value, err := response.Value.Get(value)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Received: %v -> %v\n", input, value)
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

	backoff := backoff.Fixed(time.Second, 5, 0)
	invoker := invoke.NewDynamicInvoker(staticInvoker, backoff, registry)

	time.Sleep(time.Second)
	go toLowerCase(&invoker, "Hello")
	go toLowerCase(&invoker, "World")

	time.Sleep(15 * time.Second)
	registry.Stop()
	transport.Stop()
}
