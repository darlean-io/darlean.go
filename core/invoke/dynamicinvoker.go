package invoke

import (
	"core"
	"core/backoff"
	"core/services/actorregistry"
	"core/variant"
	"math/rand"
	"time"
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

func (invoker *DynamicInvoker) Invoke(request *InvokeRequest) (any, *core.ActionError) {
	var bo backoff.BackOffSession
	useCache := true
	cacheInvalidated := false
	lazy := false
	suggestions := []string{}
	causes := []*core.ActionError{}
	var cachePreparedKey [8]byte
	triesLeft := 10
	for {
		triesLeft--
		if triesLeft <= 0 {
			break
		}

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

			if response.Error != nil {
				var error core.ActionError
				err := variant.Assign(response.Error, &error)
				if err != nil {
					causes = append(causes, core.NewFrameworkError(core.ActionErrorOptions{
						Code:     "ERROR_PARSE_ERROR",
						Template: "Action returned an error, but unable to parse the error",
					}))
				}
				if error.Kind != core.ERROR_KIND_FRAMEWORK {
					return nil, &error
				} else {
					causes = append(causes, &error)
				}

				redirect, present := error.Parameters[FRAMEWORK_ERROR_PARAMETER_REDIRECT_DESTINATION]
				if present {
					var redirects []string
					err := variant.Assign(redirect, &redirects)
					if err != nil {
						suggestions = redirects
					}
				}
				continue
				// DONE: Fill suggestions based on redirect info in error and set doBackoff to false
				// TODO: Also do this when lazy = true and other side indicates a refusal
			}

			if info.Placement.Sticky != nil && *info.Placement.Sticky {
				invoker.cache.Update(request.ActorType, request.ActorId, *receiver)
			}
			return response.Value, nil
		} else {
			causes = append(causes, core.NewFrameworkError(core.ActionErrorOptions{
				Code:     FRAMEWORK_ERROR_NO_RECEIVERS_AVAILABLE,
				Template: "No receivers available at [RequestTime] to process an action on an instance of [ActorType]",
				Parameters: map[string]any{
					"RequestTime": time.Now().UTC(),
					"ActorType":   request.ActorType,
					"ActionName":  request.ActionName,
				},
			}))
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

	var cause string
	if len(causes) > 0 {
		cause = causes[0].Message
	}

	return nil, core.NewFrameworkError(core.ActionErrorOptions{
		Code:     FRAMEWORK_ERROR_INVOKE_ERROR,
		Template: "Failed to invoke remote method [ActionName] on an instance of [ActorType]: [FirstMessage]",
		Parameters: map[string]any{
			"ActorType":    request.ActorType,
			"ActionName":   request.ActionName,
			"FirstMessage": cause,
		},
		Nested: causes,
	})
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
