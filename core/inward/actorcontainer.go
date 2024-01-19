package inward

import (
	"strconv"
	"strings"
	"sync"

	"github.com/darlean-io/darlean.go/base/actionerror"
	"github.com/darlean-io/darlean.go/core/normalized"
	"github.com/darlean-io/darlean.go/core/wire"
)

type key string

type InstanceRunner interface {
	Invoke(call *wire.ActorCallRequestIn, onFinished FinishedHandler)
	TriggerDeactivate()
}

type WrapperFactory func(id []string) InstanceWrapper

type StandardActorContainer struct {
	actorType      normalized.ActorType
	instances      map[key]InstanceRunner
	requiresLock   bool
	actionDefs     map[normalized.ActionName]ActionDef
	lock           sync.RWMutex
	wrapperFactory WrapperFactory
	onFinished     func()
	active         bool
}

func NewStandardActorContainer(actorType normalized.ActorType, requiresLock bool, actionDefs map[normalized.ActionName]ActionDef, wrapperFactory WrapperFactory, onFinished func()) *StandardActorContainer {
	return &StandardActorContainer{
		actorType:      actorType,
		instances:      make(map[key]InstanceRunner),
		requiresLock:   requiresLock,
		actionDefs:     actionDefs,
		wrapperFactory: wrapperFactory,
		onFinished:     onFinished,
		active:         true,
	}
}

func (container *StandardActorContainer) Dispatch(call *wire.ActorCallRequestIn, onFinished FinishedHandler) {
	instancerunner, err := container.obtainInstanceRunner(call.ActorId)
	if err != nil {
		onFinished(nil, err)
		return
	}
	instancerunner.Invoke(call, onFinished)
}

func (container *StandardActorContainer) obtainInstanceRunner(actorId []string) (InstanceRunner, *actionerror.Error) {
	// TODO: Only obtain write lock when item is not yet present (use read lock otherwise)
	// TODO: Do not put creation of instance runner within the lock, unless it is for the
	// same id. Different id's can be handled in parallel.
	container.lock.Lock()
	defer container.lock.Unlock()

	if !container.active {
		return nil, actionerror.NewFrameworkError(actionerror.Options{
			Code:     "CONTAINER_DEACTIVATING",
			Template: "Container is deactivating",
		})
	}

	k := makeKey(actorId)

	instancerunner, has := container.instances[k]
	if !has {
		wrapper := container.wrapperFactory(actorId)
		instancerunner = NewInstanceRunner(wrapper, container.actorType, actorId, container.requiresLock, container.actionDefs, func() {
			container.handleDeactivate(k)
		})
		container.instances[k] = instancerunner
	}

	return instancerunner, nil
}

func (container *StandardActorContainer) Stop() {
	container.lock.Lock()
	defer container.lock.Unlock()
	if !container.active {
		return
	}
	container.active = false
	for _, instance := range container.instances {
		go instance.TriggerDeactivate()
	}

	if len(container.instances) == 0 {
		if container.onFinished != nil {
			container.onFinished()
		}
	}
}

func (container *StandardActorContainer) handleDeactivate(key key) {
	container.lock.Lock()
	defer container.lock.Unlock()
	delete(container.instances, key)
	if (!container.active) && len(container.instances) == 0 {
		if container.onFinished != nil {
			container.onFinished()
		}
	}
}

func makeKey(keyParts []string) key {
	parts := make([]string, len(keyParts)*2)
	for i, p := range keyParts {
		parts[2*i] = strconv.FormatInt(int64(len(p)), 10)
		parts[2*i+1] = p
	}
	return key(strings.Join(parts, ":"))
}
