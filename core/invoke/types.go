package invoke

import (
	"core/variant"
)

type InvokeRequest struct {
	ActorType  string
	ActorId    []string
	ActionName string
	Parameters []variant.Variant
	Lazy       bool
}

type InvokeResponse struct {
	Error variant.Variant
	Value variant.Variant
}
