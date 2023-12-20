package inward

import (
	"core/normalized"
	"core/wire"
	"fmt"
)

type ActorContainer interface {
	Dispatch(call wire.Tags) error
}

type ActorInfo struct {
	actorType normalized.ActorType
	container ActorContainer
}

type Dispatcher struct {
	actorTypes map[normalized.ActorType]ActorInfo
}

func (dispatcher Dispatcher) Dispatch(call wire.Tags) error {
	actorType := call.ActorType
	if actorType == "" {
		return fmt.Errorf("Actor type not specified: %s", actorType)
	}

	normalizedActorType := normalized.NormalizeActorType(actorType)
	info, has := dispatcher.actorTypes[normalizedActorType]
	if !has {
		return fmt.Errorf("Actor type not registered: %s", actorType)
	}

	return info.container.Dispatch(call)
}

func (dispatcher Dispatcher) RegisterActorType(info ActorInfo) {
	dispatcher.actorTypes[info.actorType] = info
	dispatcher.TriggerBroadcast()
}

func (dispatcher Dispatcher) TriggerBroadcast() {
	// TODO
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}
