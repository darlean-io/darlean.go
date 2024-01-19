package inward

import (
	"github.com/darlean-io/darlean.go/core/normalized"
	"github.com/darlean-io/darlean.go/core/wire"

	"github.com/darlean-io/darlean.go/base/actionerror"
	"github.com/darlean-io/darlean.go/base/services/actorregistry"
)

type ActorContainer interface {
	Dispatch(call *wire.ActorCallRequestIn, onFinished FinishedHandler)
}

type ActorInfo struct {
	ActorType        normalized.ActorType
	Container        ActorContainer
	Placement        actorregistry.ActorPlacement
	MigrationVersion string
}

type Dispatcher struct {
	actorTypes     map[normalized.ActorType]ActorInfo
	registryPusher actorregistry.ActorRegistryPusher
}

func (dispatcher Dispatcher) Dispatch(call *wire.ActorCallRequestIn, onFinished func(*wire.ActorCallResponseOut)) {
	dispatcher.doDispatch(call, func(result any, err *actionerror.Error) {
		if err != nil {
			onFinished(&wire.ActorCallResponseOut{
				Error: err,
			})
			return
		}

		onFinished(&wire.ActorCallResponseOut{
			Value: result,
		})
	})
}

func (dispatcher Dispatcher) doDispatch(call *wire.ActorCallRequestIn, onFinished FinishedHandler) {
	actorType := call.ActorType
	if actorType == "" {
		onFinished(nil, actionerror.NewFrameworkError(actionerror.Options{
			Code:     "NO_ACTOR_TYPE",
			Template: "Actor type not specified in actor call request",
		}))
		return
	}

	normalizedActorType := normalized.NormalizeActorType(actorType)
	info, has := dispatcher.actorTypes[normalizedActorType]
	if !has {
		onFinished(nil, actionerror.NewFrameworkError(actionerror.Options{
			Code:     "ACTOR_TYPE_NOT_REGISTERED",
			Template: "Actor type [ActorType] is not registered",
			Parameters: map[string]any{
				"ActorType": actorType,
			}}))
		return
	}

	info.Container.Dispatch(call, onFinished)
}

func (dispatcher Dispatcher) RegisterActorType(info ActorInfo) {
	dispatcher.actorTypes[info.ActorType] = info
	dispatcher.TriggerBroadcast()
}

func (dispatcher Dispatcher) TriggerBroadcast() {
	info := make(map[string]actorregistry.ActorPushInfo)

	for key, value := range dispatcher.actorTypes {
		info[string(key)] = actorregistry.ActorPushInfo{
			Placement:        value.Placement,
			MigrationVersion: value.MigrationVersion,
		}
	}
	if dispatcher.registryPusher != nil {
		dispatcher.registryPusher.Set(info)
	}
}

func NewDispatcher(registryPusher actorregistry.ActorRegistryPusher) *Dispatcher {
	return &Dispatcher{
		registryPusher: registryPusher,
		actorTypes:     make(map[normalized.ActorType]ActorInfo),
	}
}
