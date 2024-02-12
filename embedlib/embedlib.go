package main

// See: https://github.com/enobufs/go-calls-c-pointer

// #include <stdint.h>
//
// typedef uint64_t handle;
//
// typedef void (*invoke_cb)(handle, _GoString_);
// void makeInvokeCallback(handle call, _GoString_ response, invoke_cb cb);
//
// typedef void (*action_cb)(handle, handle, _GoString_);
// void callActionCallback(handle action, handle call, _GoString_ request, action_cb cb);
import "C"
import (
	"runtime/cgo"
	"strings"

	"github.com/darlean-io/darlean.go/base/actionerror"
	"github.com/darlean-io/darlean.go/base/invoker"
	"github.com/darlean-io/darlean.go/utils/variant"
	"github.com/goccy/go-json"
)

var appApis map[string]*Api = map[string]*Api{}

type Handle uint64

//export CreateApp
func CreateApp(appId string, natsAddr string, hosts string) Handle {
	hostsArray := strings.Split(hosts, ",")
	_, has := appApis[appId]
	if has {
		panic("AppId already exists")
	}
	apiInstance := NewApi(appId, natsAddr, hostsArray)
	handle := cgo.NewHandle(apiInstance)
	return Handle(handle)
}

func getApi(app Handle) *Api {
	handle := cgo.Handle(app)
	return handle.Value().(*Api)
}

//export StartApp
func StartApp(app Handle) {
	getApi(app).Start()
}

//export StopApp
func StopApp(app Handle) {
	getApi(app).Stop()
}

//export ReleaseApp
func ReleaseApp(app Handle) {
	handle := cgo.Handle(app)
	handle.Delete()
}

//export Invoke
func Invoke(app Handle, call Handle, cb C.invoke_cb, options string) {
	api := getApi(app)
	goCb := func(value variant.Assignable, error *actionerror.Error) {
		var v any
		if value != nil {
			value.AssignTo(&v)
		}
		res := InvokeActionResult{
			Value: v,
			Error: fillActionError(error),
		}
		bytes, err := json.Marshal(res)
		if err != nil {
			res = InvokeActionResult{
				Error: &ActionError{
					Code:    "JSON_ERROR",
					Message: err.Error(),
				},
			}
			bytes, err = json.Marshal(res)
			if err != nil {
				panic("Fatal json error")
			}
		}

		C.makeInvokeCallback(C.handle(call), string(bytes), cb)
	}

	var opts InvokeActionOptions
	json.Unmarshal([]byte(options), &opts)
	request := invoker.Request{
		ActorType:  opts.ActorType,
		ActorId:    opts.ActorId,
		ActionName: opts.ActionName,
		Parameters: opts.Arguments,
	}
	//func (api *Api) Invoke(request *invoker.Request, goCb invokeCb) {

	api.Invoke(&request, goCb)
}

//export RegisterActor
func RegisterActor(app Handle, info string) Handle {
	api := getApi(app)
	var options RegisterActorOptions
	json.Unmarshal([]byte(info), &options)
	actor := api.RegisterActor(options)
	handle := cgo.NewHandle(actor)
	return Handle(handle)
}

//export RegisterAction
func RegisterAction(actor Handle, info string, cb C.action_cb) Handle {
	h := cgo.Handle(actor)
	a := h.Value().(RegisteredActor)

	var options RegisterActionOptions
	json.Unmarshal([]byte(info), &options)
	var actionHandle Handle

	// The c-callback that is called when someone within Darlean wants to invoke the action
	goCb := func(call PendingCall, arguments []variant.Assignable) {
		invokeOps := PerformActionOptions{
			// TODDO: ActorId,
			ActionName: options.ActionName,
			Arguments:  []any{},
		}
		for _, arg := range arguments {
			var a any
			arg.AssignTo(&a)
			invokeOps.Arguments = append(invokeOps.Arguments, a)
		}
		bytes, err := json.Marshal(invokeOps)
		if err != nil {
			panic("Fatal json error")
		}
		callHandle := cgo.NewHandle(call)
		C.callActionCallback(C.handle(actionHandle), C.handle(callHandle), string(bytes), cb)
	}

	action := a.RegisterAction(options, goCb)
	actionHandle = Handle(cgo.NewHandle((action)))
	return Handle(actionHandle)
}

//export SubmitActionResult
func SubmitActionResult(call Handle, result string) {
	var res SubmitActionResultOptions
	err := json.Unmarshal([]byte(result), &res)
	if err != nil {
		panic("Invalid json")
	}
	handle := cgo.Handle(call)
	c := handle.Value().(PendingCall)

	// Handle the response in a goroutine to avoid blocking when we are invoked
	// in the same thread that triggered the action.
	go func() {
		c.HandleResponse(res)
		handle.Delete()
	}()
}

func fillActionError(source *actionerror.Error) *ActionError {
	if source == nil {
		return nil
	}
	result := &ActionError{
		Code:       source.Code,
		Message:    source.Message,
		Template:   source.Template,
		Stack:      source.Stack,
		Kind:       ActionErrorKind(source.Kind),
		Parameters: source.Parameters,
	}
	if len(source.Nested) > 0 {
		result.Nested = [](*ActionError){}
		for _, nested := range source.Nested {
			result.Nested = append(result.Nested, fillActionError(nested))
		}
	}
	return result
}
