package remoteactorregistry

import (
	"sync"
	"time"

	"github.com/darlean-io/darlean.go/core/invoke"

	"github.com/darlean-io/darlean.go/utils/variant"

	"github.com/darlean-io/darlean.go/base/invoker"
	"github.com/darlean-io/darlean.go/base/services/actorregistry"
)

func Obtain(inv invoke.TransportInvoker, hosts []string) (*ObtainResponse, error) {
	for _, host := range hosts {
		req := invoke.TransportHandlerInvokeRequest{
			Receiver: host,
			Request: invoker.Request{
				ActorType:  SERVICE,
				ActorId:    []string{},
				ActionName: ACTION_OBTAIN,
				Parameters: []any{ObtainRequest{}},
			},
		}
		resp := inv.Invoke(&req)
		if resp.Value != nil {
			var value ObtainResponse
			err := variant.Assign(resp.Value, &value)
			if err != nil {
				return nil, err
			}
			return &value, nil
		}
	}
	return nil, nil
}

type RemoteActorRegistryFetcher struct {
	hosts     []string
	actors    map[string](actorregistry.ActorInfo)
	nonce     string
	invoker   invoke.TransportInvoker
	mutex     *sync.RWMutex
	stop      chan bool
	force     chan bool
	lastFetch time.Time
}

func (registry *RemoteActorRegistryFetcher) fetch() {
	info, err := Obtain(registry.invoker, registry.hosts)
	if err != nil {
		return
	}

	if info == nil {
		return
	}

	if info.Nonce == registry.nonce {
		return
	}

	newMap := make(map[string](actorregistry.ActorInfo))
	for key, value := range info.ActorInfo {
		newApplications := make([]actorregistry.ApplicationInfo, len(value.Applications))
		for i, v := range value.Applications {
			newApplications[i] = actorregistry.ApplicationInfo{
				Name:             v.Name,
				MigrationVersion: v.MigrationVersion,
			}
		}
		newInfo := actorregistry.ActorInfo{
			Applications: newApplications,
			Placement: actorregistry.ActorPlacement{
				AppBindIdx: value.Placement.AppBindIdx,
				Sticky:     value.Placement.Sticky,
			},
		}
		newMap[key] = newInfo
	}
	registry.mutex.Lock()
	registry.actors = newMap
	registry.nonce = info.Nonce
	registry.mutex.Unlock()

	registry.lastFetch = time.Now()
}

func (registry *RemoteActorRegistryFetcher) forceFetch() {
	now := time.Now()
	if now.Sub(registry.lastFetch) > (100 * time.Millisecond) {
		registry.fetch()
	}
}

func (registry *RemoteActorRegistryFetcher) loop(stop <-chan bool, force <-chan bool, interval time.Duration) {
	for {
		registry.fetch()
		select {
		case <-stop:
			break
		case <-force:
			registry.forceFetch()
		case <-time.After(interval):
			registry.fetch()
		}
	}
}

func (registry *RemoteActorRegistryFetcher) Get(actorType string) *actorregistry.ActorInfo {
	registry.mutex.RLock()
	info, has := registry.actors[actorType]
	registry.mutex.RUnlock()

	if !has {
		go func() {
			registry.force <- true
		}()
	}

	return &info
}

func (registry *RemoteActorRegistryFetcher) Start() {
	registry.stop = make(chan bool)
	go registry.loop(registry.stop, registry.force, 10*time.Second)
}

func (registry *RemoteActorRegistryFetcher) Stop() {
	if registry.stop != nil {
		stop := registry.stop
		registry.stop = nil
		stop <- true
	}
}

func NewFetcher(hosts []string, invoker invoke.TransportInvoker) *RemoteActorRegistryFetcher {
	force := make(chan bool)

	var mutex sync.RWMutex

	registry := RemoteActorRegistryFetcher{
		hosts:   hosts,
		actors:  make(map[string]actorregistry.ActorInfo),
		nonce:   "",
		invoker: invoker,
		mutex:   &mutex,
		force:   force,
	}

	return &registry
}
