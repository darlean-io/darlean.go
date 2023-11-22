package invoke

import "core/anny"

type InvokeRequest struct {
	ActorType  string
	ActorId    []string
	ActionName string
	Parameters []anny.Anny
	Lazy       bool
}

type InvokeResponse struct {
	Error anny.Anny
	Value anny.Anny
}
