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

func (a *GoActorImpl) Perform(actionName normalized.ActionName, args []variant.Assignable) (result any, err error) {
	fmt.Printf("GoActorImpl received %s %v\n", actionName, args)
	arg0, err := args[0].AssignToString()
	return strings.ToUpper(arg0), err
}

type GoActor_Echo struct {
	A0_Input string
	Result   string
}

type GoActor struct {
	Echo GoActor_Echo
}

type UnexistingActor_Echo struct {
	A0_Input string
	Result   string
}

type UnexistingActor struct {
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

	backoff := backoff.Exponential(1*time.Millisecond, 6, 4.0, 0.25)
	invoker := invoke.NewDynamicInvoker(transportHandler, backoff, registryFetcher)

	transportHandler.Start()
	registryPusher.Start()
	registryFetcher.Start()

	time.Sleep(time.Second)

	actionDefs := map[normalized.ActionName]inward.ActionDef{
		normalized.NormalizeActionName("Echo"): {Locking: inward.ACTION_LOCK_EXCLUSIVE},
	}

	container := inward.NewStandardActorContainer(normalized.NormalizeActorType("GoActor"), false, actionDefs, func(id []string) inward.InstanceWrapper {
		fmt.Printf("Wrapper created\n")
		return &GoActorImpl{}
	}, nil)
	disp.RegisterActorType(inward.ActorInfo{
		ActorType: "goactor",
		Container: container,
		Placement: actorregistry.ActorPlacement{},
	})

	p := portal.New(&invoker)

	// Invoke typescript actor
	tsPortal := typedportal.ForSignature[TypescriptActor](p)
	tsActor := tsPortal.Obtain([]string{})
	tsEcho := tsActor.NewCall().Echo
	tsEcho.A0_Msg = "Hello"
	tsError := tsActor.Invoke(&tsEcho)
	fmt.Printf("Received from typescript actor: %v / %v (expected: \"hello\")\n", tsEcho.Result, tsError)

	// Invoke go actor
	goPortal := typedportal.ForSignature[GoActor](p)
	goActor := goPortal.Obtain([]string{})
	goEcho := goActor.NewCall().Echo
	goEcho.A0_Input = "Foo"
	goError := goActor.Invoke(&goEcho)
	fmt.Printf("Received from go actor: %v / %+v (expected: \"FOO\")\n", goEcho.Result, goError)

	// Invoke unexisting actor type
	unexistingPortal := typedportal.ForSignature[UnexistingActor](p)
	unexistingActor := unexistingPortal.Obtain([]string{})
	unexistingEcho := unexistingActor.NewCall().Echo
	unexistingError := unexistingActor.Invoke(&unexistingEcho)
	fmt.Printf("Received from unexisting actor: %v / %+v (expected: \"<error>\")\n", unexistingEcho.Result, unexistingError)

	time.Sleep(time.Second)
	go toLowerCase(&invoker, "Hello")
	go toLowerCase(&invoker, "World")

	time.Sleep(15 * time.Second)
	registryFetcher.Stop()
	registryPusher.Stop()
	transport.Stop()
}
