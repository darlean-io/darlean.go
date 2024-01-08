package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/darlean-io/darlean.go/base/invoker"
	"github.com/darlean-io/darlean.go/base/portal"
	"github.com/darlean-io/darlean.go/base/services/actorregistry"
	"github.com/darlean-io/darlean.go/base/typedportal"
	"github.com/darlean-io/darlean.go/core/backoff"
	"github.com/darlean-io/darlean.go/core/invoke"
	"github.com/darlean-io/darlean.go/core/inward"
	"github.com/darlean-io/darlean.go/core/natstransport"
	"github.com/darlean-io/darlean.go/core/normalized"
	"github.com/darlean-io/darlean.go/core/remoteactorregistry"
	"github.com/darlean-io/darlean.go/core/transporthandler"
	"github.com/darlean-io/darlean.go/utils/variant"
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

func toLowerCase(inv *invoke.DynamicInvoker, input string) {
	req := invoker.Request{
		ActorType:  "echoactor",
		ActorId:    []string{"A"},
		ActionName: "echo",
		Parameters: []any{input},
	}

	value, err := inv.Invoke(&req)
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

type TypescriptActor_Echo struct {
	A0_Msg string
	Result string
}

type TypescriptActor struct {
	Echo TypescriptActor_Echo
}

type GoActorImpl struct{}

func (a *GoActorImpl) Activate() error {
	return nil
}

func (a *GoActorImpl) Deactivate() error {
	return nil
}

func (a *GoActorImpl) Perform(actionName normalized.ActionName, args []any) (result any, err error) {
	fmt.Printf("GoActorImpl received %s %v\n", actionName, args)
	return strings.ToUpper(args[0].(string)), nil
}

type GoActor_Echo struct {
	A0_Input string
	Result   string
}

type GoActor struct {
	Echo GoActor_Echo
}

func main() {

	const OUR_APP_ID = "client"
	const NATS_ADDR = "localhost:4500"
	HOSTS := []string{"server"}

	transport, err := natstransport.New(NATS_ADDR, OUR_APP_ID)
	if err != nil {
		panic(err)
	}

	var disp *inward.Dispatcher

	transportHandler := transporthandler.New(transport, func() transporthandler.InwardCallDispatcher {
		return disp
	}, OUR_APP_ID)
	registryFetcher := remoteactorregistry.NewFetcher(HOSTS, transportHandler)
	registryPusher := remoteactorregistry.NewPusher(HOSTS, OUR_APP_ID, transportHandler)
	disp = inward.NewDispatcher(registryPusher)

	backoff := backoff.Exponential(10*time.Millisecond, 8, 4.0, 0.25)
	invoker := invoke.NewDynamicInvoker(transportHandler, backoff, registryFetcher)

	transportHandler.Start()
	registryPusher.Start()
	registryFetcher.Start()

	time.Sleep(time.Second)

	container := inward.NewStandardActorContainer(false, map[normalized.ActionName]inward.ActionDef{}, func(id []string) inward.InstanceWrapper {
		fmt.Printf("Wrapper created\n")
		return &GoActorImpl{}
	}, nil)
	disp.RegisterActorType(inward.ActorInfo{
		ActorType: "goactor",
		Container: container,
		Placement: actorregistry.ActorPlacement{},
	})

	p := portal.New(&invoker)

	tsPortal := typedportal.ForSignature[TypescriptActor](p)
	tsActor := tsPortal.Obtain([]string{})
	tsEcho := tsActor.NewCall().Echo
	tsEcho.A0_Msg = "Hello"
	tsError := tsActor.Invoke(&tsEcho)
	fmt.Printf("Received from typescript actor: %v %v\n", tsEcho.Result, tsError)

	goPortal := typedportal.ForSignature[GoActor](p)
	goActor := goPortal.Obtain([]string{})
	goEcho := goActor.NewCall().Echo
	goEcho.A0_Input = "Foo"
	goError := goActor.Invoke(&goEcho)
	fmt.Printf("Received from go actor: %v / %v\n", goEcho.Result, goError)

	/*echoPortal := typedportal.ForSignature[EchoActor](p)
	actor := echoPortal.Obtain([]string{"abc"})
	call := actor.NewCall().Echo
	call.A0 = "Hello"
	call.A2 = 42
	err = actor.Invoke(&call)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Received via Portal: %v\n", call.Result)
	*/

	time.Sleep(time.Second)
	go toLowerCase(&invoker, "Hello")
	go toLowerCase(&invoker, "World")

	time.Sleep(15 * time.Second)
	registryFetcher.Stop()
	registryPusher.Stop()
	transport.Stop()
}
