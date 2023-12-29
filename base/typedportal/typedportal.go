package typedportal

import (
	"github.com/darlean-io/darlean.go/base/portal"
	baseportal "github.com/darlean-io/darlean.go/base/portal"
)

// Interface to a portal that returns proxies of a specific actor type.
type Portal[ActorSig baseportal.ActorSignature] interface {
	Obtain(id []string) *portal.ActorProxy[ActorSig]
}

// typedPortal implements [typedportal.Portal]
type typedPortal[ActorSig baseportal.ActorSignature] struct {
	base baseportal.Portal
}

func (portal typedPortal[ActorSig]) Obtain(id []string) *portal.ActorProxy[ActorSig] {
	return &baseportal.ActorProxy[ActorSig]{
		Base: portal.base,
		Id:   id,
	}
}

// Returns a typed portal for the type of the provided actor signature that uses the base portal to
// actually invoke actions.
func ForSignature[ActorSig baseportal.ActorSignature](basePortal portal.Portal) Portal[ActorSig] {
	p := typedPortal[ActorSig]{
		base: basePortal,
	}
	return p
}
