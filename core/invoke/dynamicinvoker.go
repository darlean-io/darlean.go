package invoke

import (
	"core/backoff"
	"core/services/actorregistry"
	"core/variant"
	"math/rand"
)

type DynamicInvoker struct {
	staticInvoker *StaticInvoker
	backoff       backoff.BackOff
	registry      actorregistry.ActorRegistry
	cache         *PlacementCache
}

func NewDynamicInvoker(staticInvoker *StaticInvoker, backoff backoff.BackOff, registry actorregistry.ActorRegistry) DynamicInvoker {
	return DynamicInvoker{
		staticInvoker: staticInvoker,
		backoff:       backoff,
		registry:      registry,
		cache:         NewPlacementCache(),
	}
}

func (invoker *DynamicInvoker) Invoke(request *InvokeRequest) *InvokeResponse {
	var bo backoff.BackOffSession
	useCache := true
	cacheInvalidated := false
	lazy := false
	suggestions := []string{}
	var cachePreparedKey [8]byte
	for {
		info := invoker.registry.Get(request.ActorType)
		receiver := extractBindName(request.ActorId, info.Placement.AppBindIdx)
		doBackoff := true

		if info.Placement.Sticky != nil && *info.Placement.Sticky {
			if useCache {
				cachePreparedKey = invoker.cache.Prepare(request.ActorType, request.ActorId)
				receiver = invoker.cache.Get(cachePreparedKey)
				// Only use the cache on the zero-th retry. When a retry is necessary,
				// we cannot trust the cache.
				useCache = false
				lazy = true
			} else {
				if !cacheInvalidated {
					invoker.cache.Delete(cachePreparedKey)
					cacheInvalidated = true
				}
			}
		}

		var appIdx = -1

		var applications []string
		if len(suggestions) > 0 {
			applications = suggestions
		} else {
			applications = make([]string, len(info.Applications))
			for i, app := range info.Applications {
				applications[i] = app.Name
			}
		}

		if receiver == nil {
			switch len(applications) {
			case 0:
				break
			case 1:
				receiver = &applications[0]
			default:
				if appIdx < 0 {
					appIdx = rand.Intn(len(applications))
				} else {
					appIdx++
				}
				idx := appIdx % len(applications)
				receiver = &applications[idx]
			}
		}

		if receiver != nil {
			staticRequest := StaticInvokeRequest{
				InvokeRequest: *request,
				Receiver:      *receiver,
			}
			staticRequest.Lazy = lazy
			lazy = false
			response := invoker.staticInvoker.Invoke(&staticRequest)
			if response.Error == nil {
				return response
			}
			// TODO: Fill suggestions based on redirect info in error and set doBackoff to false
			// Also do this when lazy = true and other side indicates a refusal
			// Do something with MigrationVersion?
		}

		if !doBackoff {
			continue
		}

		if bo == nil {
			bo = invoker.backoff.Begin()
		}

		if !bo.BackOff() {
			break
		}
	}

	return &InvokeResponse{
		Error: variant.New("invoke: All attempts failed"),
	}
}

func extractBindName(id []string, bindIdx *int) *string {
	var idx int
	if bindIdx == nil {
		return nil
	}

	if *bindIdx < 0 {
		idx = len(id) + *bindIdx
	} else {
		idx = *bindIdx
	}

	if idx < 0 || idx >= len(id) {
		return nil
	}
	return &id[idx]
}
