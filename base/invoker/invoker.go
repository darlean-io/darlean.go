/*
Package invoker defines the interface for invoking remote actors.
*/
package invoker

import (
	"github.com/darlean-io/darlean.go/base/actionerror"
	"github.com/darlean-io/darlean.go/utils/variant"
)

/*
Request contains the fields for invoking a remote action.
*/
type Request struct {
	ActorType  string
	ActorId    []string
	ActionName string
	Parameters []any
	Lazy       bool
}

/*
Response contains the results of invoking a remote action.
*/
type Response struct {
	Error variant.Assignable
	Value variant.Assignable
}

// Invoker can invoke remote actors.
type Invoker interface {
	// Invoke performs the request and returns the result value or error.
	Invoke(request *Request) (variant.Assignable, *actionerror.Error)
}
