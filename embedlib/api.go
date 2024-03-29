package main

// Note: We use uint64 as handles (instead of uintptr) because that allows future implementations
// to put something else than a pointer inside (like an every incrementing sequence number). This may be
// relevant for non-64-bit platforms.

// #include <stdint.h>
//
// typedef uint64_t handle;
//
// typedef void (*invoke_cb)(handle, _GoString_);
// extern void makeInvokeCallback(handle call, _GoString_ response, invoke_cb cb) {
//     cb(call, response);
// }
//
// typedef void (*action_cb)(handle, handle, _GoString_);
// extern void callActionCallback(handle action, handle call, _GoString_ request, action_cb cb) {
//     cb(action, call, request);
//}
import "C"

import (
	"time"

	"github.com/darlean-io/darlean.go/base/actionerror"
	"github.com/darlean-io/darlean.go/base/invoker"
	"github.com/darlean-io/darlean.go/base/services/actorregistry"
	"github.com/darlean-io/darlean.go/core/backoff"
	"github.com/darlean-io/darlean.go/core/invoke"
	"github.com/darlean-io/darlean.go/core/inward"
	"github.com/darlean-io/darlean.go/core/natstransport"
	"github.com/darlean-io/darlean.go/core/normalized"
	"github.com/darlean-io/darlean.go/core/remoteactorregistry"
	"github.com/darlean-io/darlean.go/core/transporthandler"
	"github.com/darlean-io/darlean.go/utils/variant"
)

type invokeCb func(value variant.Assignable, error *actionerror.Error)
type actionCb func(call PendingCall, arguments []variant.Assignable)

type ActionInfo struct {
	ActionName normalized.ActionName
	Locking    inward.ActionLockKind
	Callback   actionCb
}

type ActorInfo struct {
	ActorType   normalized.ActorType
	Actions     map[normalized.ActionName]ActionInfo
	CallManager CallManager
}

type CallManager interface {
	MakeActionCall(info *ActionInfo, resultChannel chan SubmitActionResultOptions, arguments []variant.Assignable)
}

type CallInfo struct {
	resultChannel chan SubmitActionResultOptions
}

type Api struct {
	Invoker       *invoke.DynamicInvoker
	registry      *remoteactorregistry.RemoteActorRegistryFetcher
	transport     *natstransport.NatsTransport
	staticInvoker *transporthandler.TransportHandler
	fetcher       *remoteactorregistry.RemoteActorRegistryFetcher
	actorTypes    map[normalized.ActorType]ActorInfo
	pusher        *remoteactorregistry.RemoteActorRegistryPusher
	dispatcher    *inward.Dispatcher
	containers    []*inward.StandardActorContainer
}

type RegisteredActor interface {
	RegisterAction(options RegisterActionOptions, callback actionCb) RegisteredAction
}

type RegisteredAction interface {
}

type PendingCall interface {
	HandleResponse(options SubmitActionResultOptions)
}

func NewApi(appId string, natsAddr string, hosts []string) *Api {
	transport, err := natstransport.New(natsAddr, appId)
	if err != nil {
		panic(err)
	}

	staticInvoker := transporthandler.New(transport, appId)
	fetcher := remoteactorregistry.NewFetcher(hosts, staticInvoker)

	backoff := backoff.Exponential(1*time.Millisecond, 6, 4.0, 0.25)
	invoker := invoke.NewDynamicInvoker(staticInvoker, backoff, fetcher)

	registryPusher := remoteactorregistry.NewPusher(hosts, appId, staticInvoker)
	dispatcher := inward.NewDispatcher(registryPusher)

	return &Api{
		Invoker:       &invoker,
		registry:      fetcher,
		transport:     transport,
		staticInvoker: staticInvoker,
		fetcher:       fetcher,
		pusher:        registryPusher,
		dispatcher:    dispatcher,
		actorTypes:    map[normalized.ActorType]ActorInfo{},
		containers:    []*inward.StandardActorContainer{},
	}
}

func (api *Api) Start() {
	api.registerActors()
	api.staticInvoker.Start(api.dispatcher)
	api.fetcher.Start()
	if len(api.actorTypes) > 0 {
		api.pusher.Start()
	}
}

func (api *Api) Stop() {
	if len(api.actorTypes) > 0 {
		for idx := len(api.containers) - 1; idx >= 0; idx-- {
			api.containers[idx].Stop()
		}
		api.pusher.Stop()
	}
	api.registry.Stop()
	api.transport.Stop()
}

func (api *Api) Invoke(request *invoker.Request, goCb invokeCb) {
	result, err := api.Invoker.Invoke(request)
	goCb(result, err)
}

func (actor *ActorInfo) RegisterAction(options RegisterActionOptions, callback actionCb) RegisteredAction {
	normalizedActionName := normalized.NormalizeActionName(options.ActionName)
	var actionLocking inward.ActionLockKind
	if options.Locking == "shared" {
		actionLocking = inward.ACTION_LOCK_SHARED
	} else if options.Locking == "none" {
		actionLocking = inward.ACTION_LOCK_NONE
	} else {
		actionLocking = inward.ACTION_LOCK_EXCLUSIVE
	}

	action := ActionInfo{
		ActionName: normalizedActionName,
		Locking:    actionLocking,
		Callback:   callback,
	}
	actor.Actions[normalizedActionName] = action
	return &action
}

func (api *Api) RegisterActor(options RegisterActorOptions) RegisteredActor {
	normalizedActorType := normalized.NormalizeActorType(options.ActorType)
	info := ActorInfo{
		ActorType:   normalizedActorType,
		Actions:     map[normalized.ActionName]ActionInfo{},
		CallManager: api,
	}

	api.actorTypes[normalizedActorType] = info
	return &info
}

func (api *Api) MakeActionCall(info *ActionInfo, resultChannel chan SubmitActionResultOptions, arguments []variant.Assignable) {
	call := CallInfo{
		resultChannel: resultChannel,
	}
	info.Callback(&call, arguments)
}

func (call *CallInfo) HandleResponse(options SubmitActionResultOptions) {
	call.resultChannel <- options
}

func (api *Api) registerActors() {
	if len(api.actorTypes) == 0 {
		return
	}

	for _, actor2 := range api.actorTypes {
		actor := actor2 // Make a scoped copy so that coroutines have the proper actor
		actionDefs := map[normalized.ActionName]inward.ActionDef{}
		for _, action := range actor.Actions {
			actionDefs[action.ActionName] = inward.ActionDef{
				Locking: action.Locking,
			}
		}

		// TODO: fix RequiresLock
		container := inward.NewStandardActorContainer(actor.ActorType, false, actionDefs,
			// Wrapper factory:
			func(id []string) inward.InstanceWrapper {

				stub := NewActorStub(&actor, id)
				return stub
			},
			// Container finished handler:
			func() {

			})
		api.containers = append(api.containers, container)

		api.dispatcher.RegisterActorType(inward.ActorInfo{
			ActorType: actor.ActorType,
			Container: container,
			Placement: actorregistry.ActorPlacement{},
		})
	}
}

func main() {}
