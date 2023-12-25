package remoteactorregistry

import (
	"fmt"
	"time"

	"github.com/darlean-io/darlean.go/core/invoke"

	"github.com/darlean-io/darlean.go/base/services/actorregistry"
)

func Push(invoker invoke.TransportInvoker, hosts []string, request PushRequest) error {
	for _, host := range hosts {
		fmt.Sprintf("Pushing to %v", host)
		req := invoke.TransportHandlerInvokeRequest{
			Receiver: host,
			InvokeRequest: invoke.InvokeRequest{
				ActorType:  SERVICE,
				ActorId:    []string{},
				ActionName: ACTION_PUSH,
				Parameters: []any{request},
			},
		}
		resp := invoker.Invoke(&req)
		if resp.Error == nil {
			return nil
		}
	}
	return nil
}

type RemoteActorRegistryPusher struct {
	appId    string
	hosts    []string
	info     map[string]ActorPushInfo
	invoker  invoke.TransportInvoker
	stop     chan bool
	force    chan bool
	lastPush time.Time
}

func (registry *RemoteActorRegistryPusher) push() {
	registry.lastPush = time.Now()

	if registry.info == nil {
		return
	}

	Push(registry.invoker, registry.hosts, PushRequest{
		Application: registry.appId,
		ActorInfo:   registry.info,
	})
}

func (registry *RemoteActorRegistryPusher) forcePush() {
	now := time.Now()
	if now.Sub(registry.lastPush) > (100 * time.Millisecond) {
		registry.push()
	}
}

func (registry *RemoteActorRegistryPusher) loop(stop <-chan bool, force <-chan bool, interval time.Duration) {
	for {
		registry.push()
		select {
		case <-stop:
			break
		case <-force:
			registry.forcePush()
		case <-time.After(interval):
			registry.push()
		}
	}
}

func (registry *RemoteActorRegistryPusher) Set(info map[string]actorregistry.ActorPushInfo) {
	registry.info = map[string]ActorPushInfo{}
	for key, value := range info {
		registry.info[key] = ActorPushInfo{
			Placement:        ActorPlacement(value.Placement),
			MigrationVersion: value.MigrationVersion,
		}
	}
	go func() {
		registry.force <- true
	}()
}

func (registry *RemoteActorRegistryPusher) Stop() {
	registry.stop <- true
}

func NewPusher(hosts []string, appId string, invoker invoke.TransportInvoker) *RemoteActorRegistryPusher {
	stop := make(chan bool)
	force := make(chan bool)

	registry := RemoteActorRegistryPusher{
		hosts:   hosts,
		appId:   appId,
		invoker: invoker,
		stop:    stop,
		force:   force,
	}

	go registry.loop(stop, force, 10*time.Second)

	return &registry
}
