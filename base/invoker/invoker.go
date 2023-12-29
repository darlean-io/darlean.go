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

type Invoker interface {
	Invoke(request *Request) (any, *base.ActionError)
}
