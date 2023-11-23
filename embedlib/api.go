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
	"time"
)

type Api struct {
	Invoker   *invoke.DynamicInvoker
	registry  *remoteactorregistry.RemoteActorRegistry
	transport *natstransport.NatsTransport
}

func NewApi(appId string, natsAddr string, hosts []string) *Api {
	transport, err := natstransport.New(natsAddr, appId)
	if err != nil {
		panic(err)
	}

	staticInvoker := invoke.NewStaticInvoker(transport, appId)
	registry := remoteactorregistry.New(hosts, staticInvoker)

	backoff := backoff.Fixed(time.Second, 8, 0)
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
	response := api.Invoker.Invoke(request)

	if response.Value != nil {
		var value any
		value, err := response.Value.Get(value)
		if err != nil {
			panic(err)
		}
		goCb(value.(string))
	}
}

func main() {}
