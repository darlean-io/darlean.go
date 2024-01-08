/*
Package invoker defines the interface for invoking remote actors.
*/
package invoker

import "github.com/darlean-io/darlean.go/base"

type Request struct {
	ActorType  string
	ActorId    []string
	ActionName string
	Parameters []any
	Lazy       bool
}

type Response struct {
	Error any
	Value any
}

// Invoker can invoke remote actors.
type Invoker interface {
	// Invoke performs the request and returns the result value or error.
	Invoke(request *Request) (any, *base.ActionError)
}
