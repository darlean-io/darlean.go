package main

// typedef void (*invoke_cb)(int);
// extern void makeCallback(int bufHandle, invoke_cb cb) {
//     cb(bufHandle);
// }

// typedef void (*invoke_cb)(_GoString_);
// extern void makeCallback(_GoString_ bufHandle, invoke_cb cb) {
//     cb(bufHandle);
// }
import "C"

import (
	"core/backoff"
	"core/invoke"
	"core/natstransport"
	"core/remoteactorregistry"
	"core/transporthandler"
	"core/variant"
	"fmt"
	"time"
)

type Api struct {
	Invoker   *invoke.DynamicInvoker
	registry  *remoteactorregistry.RemoteActorRegistryFetcher
	transport *natstransport.NatsTransport
}

func NewApi(appId string, natsAddr string, hosts []string) *Api {
	transport, err := natstransport.New(natsAddr, appId)
	if err != nil {
		panic(err)
	}

	staticInvoker := transporthandler.New(transport, nil, appId)
	registry := remoteactorregistry.NewFetcher(hosts, staticInvoker)

	backoff := backoff.Exponential(10*time.Millisecond, 8, 4.0, 0.25)
	invoker := invoke.NewDynamicInvoker(staticInvoker, backoff, registry)

	return &Api{
		Invoker:   &invoker,
		registry:  registry,
		transport: transport,
	}
}

func (api Api) Stop() {
	api.registry.Stop()
	api.transport.Stop()
}

type invokeCb func(n string)

func (api Api) Invoke(request *invoke.InvokeRequest, goCb invokeCb) {
	result, err := api.Invoker.Invoke(request)

	if err != nil {
		goCb("invoke: " + err.Error())
	}

	// TODO encode variant
	var value any
	variant.Assign(result, &value)
	goCb(fmt.Sprint(value))
}

func main() {}
