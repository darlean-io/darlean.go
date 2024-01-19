/*
Package typedportal provides an interface to and an implementation of a portal that
provides access to an actor of a specific type.

It makes heavy use of the concepts of [signature.Actor] and [signature.Action] to
invoke remote actors in a type-safe way.

Use [ForSignature] to obtain a typed portal from a base portal for a given actor signature.
*/
package typedportal

import (
	"github.com/darlean-io/darlean.go/base/portal"
	"github.com/darlean-io/darlean.go/base/signature"
)

// Interface to a typed portal that returns proxies of a specific actor type.
type Portal[ActorSig signature.Actor] interface {
	// Returns a new proxy for the provided id.
	Obtain(id []string) *portal.ActorProxy[ActorSig]
}

// typedPortal satisfies [typedportal.Portal]
type typedPortal[ActorSig signature.Actor] struct {
	base portal.Portal
}

// Obtain satisfies [typedportal.Portal].
func (p typedPortal[ActorSig]) Obtain(id []string) *portal.ActorProxy[ActorSig] {
	return &portal.ActorProxy[ActorSig]{
		Base: p.base,
		Id:   id,
	}
}

// Returns a typed portal for the type of the provided actor signature that uses the base portal to
// actually invoke actions.
func ForSignature[ActorSig signature.Actor](basePortal portal.Portal) Portal[ActorSig] {
	p := typedPortal[ActorSig]{
		base: basePortal,
	}
	return p
}
