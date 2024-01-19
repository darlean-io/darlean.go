package remoteactorregistry

import (
	"time"

	"github.com/darlean-io/darlean.go/core/invoke"

	"github.com/darlean-io/darlean.go/base/invoker"
	"github.com/darlean-io/darlean.go/base/services/actorregistry"
)

func Push(inv invoke.TransportInvoker, hosts []string, request PushRequest) error {
	for _, host := range hosts {
		// fmt.Printf("Pushing to %v: %+v\n", host, request)
		req := invoke.TransportHandlerInvokeRequest{
			Receiver: host,
			Request: invoker.Request{
				ActorType:  SERVICE,
				ActorId:    []string{},
				ActionName: ACTION_PUSH,
				Parameters: []any{request},
			},
		}
		resp := inv.Invoke(&req)
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

func (registry *RemoteActorRegistryPusher) Start() {
	registry.stop = make(chan bool)
	go registry.loop(registry.stop, registry.force, 10*time.Second)
}

func (registry *RemoteActorRegistryPusher) Stop() {
	if registry.stop != nil {
		stop := registry.stop
		registry.stop = nil
		stop <- true
	}
}

func NewPusher(hosts []string, appId string, invoker invoke.TransportInvoker) *RemoteActorRegistryPusher {
	force := make(chan bool)

	registry := RemoteActorRegistryPusher{
		hosts:   hosts,
		appId:   appId,
		invoker: invoker,
		force:   force,
	}

	return &registry
}
