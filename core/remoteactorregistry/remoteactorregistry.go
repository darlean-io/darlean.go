package remoteactorregistry

import (
	"core/invoke"
	"core/services/actorregistry"
	"core/variant"
	"sync"
	"time"
)

const SERVICE = "io.darlean.actorregistryservice"

type ObtainRequest struct {
}

type ApplicationInfo struct {
	Name             string  `json:"name"`
	MigrationVersion *string `json:"migrationVersion"`
}

type ActorPlacement struct {
	AppBindIdx *int  `json:"appBindIdx"`
	Sticky     *bool `json:"sticky"`
}

type ActorInfo struct {
	Applications []ApplicationInfo `json:"applications"`
	Placement    ActorPlacement    `json:"Placement"`
}

type ObtainResponse struct {
	Nonce     string               `json:"nonce"`
	ActorInfo map[string]ActorInfo `json:"actorInfo"`
}

const ACTION_OBTAIN = "obtain"

func Obtain(invoker *invoke.StaticInvoker, hosts []string) (*ObtainResponse, error) {
	for _, host := range hosts {
		req := invoke.StaticInvokeRequest{
			Receiver: host,
			InvokeRequest: invoke.InvokeRequest{
				ActorType:  SERVICE,
				ActorId:    []string{},
				ActionName: ACTION_OBTAIN,
				Parameters: []any{ObtainRequest{}},
			},
		}
		resp := invoker.Invoke(&req)
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

type RemoteActorRegistry struct {
	hosts     []string
	actors    map[string](actorregistry.ActorInfo)
	nonce     string
	invoker   *invoke.StaticInvoker
	mutex     *sync.RWMutex
	stop      chan bool
	force     chan bool
	lastFetch time.Time
}

func (registry *RemoteActorRegistry) fetch() {
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

func (registry *RemoteActorRegistry) forceFetch() {
	now := time.Now()
	if now.Sub(registry.lastFetch) > (100 * time.Millisecond) {
		registry.fetch()
	}
}

func (registry *RemoteActorRegistry) loop(stop <-chan bool, force <-chan bool, interval time.Duration) {
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

func (registry *RemoteActorRegistry) Get(actorType string) *actorregistry.ActorInfo {
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

func (registry *RemoteActorRegistry) Stop() {
	registry.stop <- true
}

func New(hosts []string, invoker *invoke.StaticInvoker) *RemoteActorRegistry {
	stop := make(chan bool)
	force := make(chan bool)

	var mutex sync.RWMutex

	registry := RemoteActorRegistry{
		hosts:   hosts,
		actors:  make(map[string]actorregistry.ActorInfo),
		nonce:   "",
		invoker: invoker,
		mutex:   &mutex,
		stop:    stop,
		force:   force,
	}

	go registry.loop(stop, force, 10*time.Second)

	return &registry
}
