package portal

import (
	"github.com/darlean-io/darlean.go/base/invoker"
)

// Portal ressembles a portal that can be used to invoke remote actors of any type or id.
// An instance of a portal can be created by means of [portal.New].
type Portal interface {
	invoker.Invoker
}

// StandardPortal is the standard implementation of a Portal. Can be constructed using [portal.New].
type invokerPortal struct {
	invoker.Invoker
}

// New creates and returns a new Portal for the provided invoker. The portal uses
// the invoker to make actual calls to actors.
func New(invoker invoker.Invoker) Portal {
	return invokerPortal{
		Invoker: invoker,
	}
}
