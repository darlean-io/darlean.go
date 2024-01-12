package portal

import (
	"reflect"
	"strings"

	"github.com/darlean-io/darlean.go/base/invoker"
	"github.com/darlean-io/darlean.go/base/signature"
	"github.com/darlean-io/darlean.go/utils/variant"
)

// ActorProxy is a proxy to a remote actor than can be used to invoke methods on a remote actor.
// ActorSig must be a valid [portal.ActorSignature].
type ActorProxy[ActorSig signature.Actor] struct {
	// Base is the portal to be used to invoke the remote actor.
	Base Portal
	// Id is the id of the remote actor to which this proxy points.
	Id []string
}

// Invoke invokes an action on the remote actor.
// The action must be a valid ActionSignature for ActorSig. The argument values (`A0`, `A1` et cetera)
// must be properly set.
// After the invocation, the `Result` field of the action contains the result (when no error occurred). Otherwise,
// the error is returned.
func (proxy ActorProxy[ActorSig]) Invoke(action signature.Action) error {
	var a ActorSig
	var tp = reflect.TypeOf(a)
	var inputtp = reflect.ValueOf(action).Elem().Type()
	var inputtps = strings.Split(inputtp.Name(), "_")
	var actionName = inputtps[len(inputtps)-1]

	req := invoker.Request{
		ActorType:  strings.ToLower(tp.Name()),
		ActorId:    proxy.Id,
		ActionName: strings.ToLower(actionName),
	}
	a0 := reflect.ValueOf(action).Elem().FieldByName("A0")
	if (a0 == reflect.Value{}) {
		a0 = reflect.ValueOf(action).Elem().FieldByNameFunc(func(name string) bool {
			return strings.HasPrefix(name, "A0_")
		})
	}
	a0value := a0.Interface()
	req.Parameters = []any{a0value}
	resp, err := proxy.Base.Invoke(&req)
	if err != nil {
		return err
	}
	res := reflect.ValueOf(action)
	res = res.Elem().FieldByName("Result")
	return variant.Assign(resp, &res)
}

// NewCall returns a new instance of Calls that can be used to make a new call.
// Domain logic can fill in one of the `A0`, `A1` argument values of the returned object,
// and pass that to [ActorProxy.Invoke], which will then invoke the action and fill in the
// `Result` value of the object.
func (proxy ActorProxy[ActorSig]) NewCall() ActorSig {
	var t ActorSig
	return t
}
