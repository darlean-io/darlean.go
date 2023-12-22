package inward

import (
	"core/normalized"
	"core/services/actorregistry"
	"core/wire"
	"fmt"
)

type ActorContainer interface {
	Dispatch(call *wire.ActorCallRequest, onFinished FinishedHandler)
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

func (dispatcher Dispatcher) Dispatch(call *wire.ActorCallRequest, onFinished func(*wire.ActorCallResponse)) {
	dispatcher.doDispatch(call, func(result any, err error) {
		if err != nil {
			onFinished(&wire.ActorCallResponse{
				Error: err,
			})
			return
		}

		onFinished(&wire.ActorCallResponse{
			Value: result,
		})
	})
}

func (dispatcher Dispatcher) doDispatch(call *wire.ActorCallRequest, onFinished FinishedHandler) {
	actorType := call.ActorType
	if actorType == "" {
		onFinished(nil, fmt.Errorf("Actor type not specified: %s", actorType))
		return
	}

	normalizedActorType := normalized.NormalizeActorType(actorType)
	info, has := dispatcher.actorTypes[normalizedActorType]
	if !has {
		onFinished(nil, fmt.Errorf("Actor type not registered: %s", actorType))
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
	}
}
