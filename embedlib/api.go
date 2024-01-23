package main

// typedef void (*invoke_cb)(_GoString_);
// extern void makeCallback(_GoString_ bufHandle, invoke_cb cb) {
//     cb(bufHandle);
// }
//
// typedef void (*action_cb)(_GoString_);
// extern void callActionCallback(_GoString_ bufHandle, action_cb cb) {
//     cb(bufHandle);
//}
import "C"

import (
	"fmt"
	"time"

	"github.com/darlean-io/darlean.go/base/invoker"
	"github.com/darlean-io/darlean.go/core/backoff"
	"github.com/darlean-io/darlean.go/core/invoke"
	"github.com/darlean-io/darlean.go/core/natstransport"
	"github.com/darlean-io/darlean.go/core/remoteactorregistry"
	"github.com/darlean-io/darlean.go/core/transporthandler"
)

type Api struct {
	Invoker       *invoke.DynamicInvoker
	registry      *remoteactorregistry.RemoteActorRegistryFetcher
	transport     *natstransport.NatsTransport
	staticInvoker *transporthandler.TransportHandler
	fetcher       *remoteactorregistry.RemoteActorRegistryFetcher
}

func NewApi(appId string, natsAddr string, hosts []string) *Api {
	transport, err := natstransport.New(natsAddr, appId)
	if err != nil {
		panic(err)
	}

	staticInvoker := transporthandler.New(transport, nil, appId)
	fetcher := remoteactorregistry.NewFetcher(hosts, staticInvoker)

	backoff := backoff.Exponential(1*time.Millisecond, 6, 4.0, 0.25)
	invoker := invoke.NewDynamicInvoker(staticInvoker, backoff, fetcher)

	return &Api{
		Invoker:       &invoker,
		registry:      fetcher,
		transport:     transport,
		staticInvoker: staticInvoker,
		fetcher:       fetcher,
	}
}

func (api *Api) Start() {
	api.staticInvoker.Start()
	api.fetcher.Start()
}

func (api *Api) Stop() {
	api.registry.Stop()
	api.transport.Stop()
}

type invokeCb func(n string)

func (api *Api) Invoke(request *invoker.Request, goCb invokeCb) {
	result, err := api.Invoker.Invoke(request)

	if err != nil {
		goCb("error: " + err.Error())
		return
	}

	// TODO encode variant as JSON
	var value any
	err2 := result.AssignTo(&value)
	if err2 != nil {
		goCb("error: " + err2.Error())
		return
	}

	goCb(fmt.Sprint(value))
}

func main() {}
