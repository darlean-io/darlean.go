package main

// See: https://github.com/enobufs/go-calls-c-pointer

// typedef void (*invoke_cb)(_GoString_);
// void makeCallback(_GoString_ bufhandle, invoke_cb cb);
//
// typedef void (*action_cb)(_GoString_);
// void callActionCallback(_GoString_ bufhandle, action_cb cb);
import "C"
import (
	"strings"

	"github.com/darlean-io/darlean.go/base/invoker"
)

var appApis map[string]*Api = map[string]*Api{}

//export CreateApp
func CreateApp(appId string, natsAddr string, hosts string) {
	hostsArray := strings.Split(hosts, ",")
	_, has := appApis[appId]
	if has {
		panic("AppId already exists")
	}
	apiInstance := NewApi(appId, natsAddr, hostsArray)
	appApis[appId] = apiInstance
}

func getApi(appId string) *Api {
	api, has := appApis[appId]
	if !has {
		panic("No app for provided appId")
	}
	return api
}

//export StartApp
func StartApp(appId string) {
	getApi(appId).Start()
}

//export StopApp
func StopApp(appId string) {
	getApi(appId).Stop()
}

//export ReleaseApp
func ReleaseApp(appId string) {
	delete(appApis, appId)
}

//export Invoke
func Invoke(appId string, cb C.invoke_cb, actorType string, actorId []string, actionName string, arguments string) {
	api := getApi(appId)
	//goCb := func(bufhandle int) {
	//	C.makeCallback(C.int(bufhandle), cb)
	//}
	goCb := func(bufhandle string) {
		C.makeCallback(bufhandle, cb)
	}
	request := invoker.Request{
		ActorType:  actorType,
		ActorId:    actorId, //strings.Split(actorId, ","),
		ActionName: actionName,
		Parameters: []any{arguments},
	}
	api.Invoke(&request, goCb)
}

//export RegisterActor
func RegisterActor(appId string, info string) {
	// TODO
}

//export RegisterAction
func RegisterAction(appId string, info string, cb C.action_cb) {
	// TODO
}

//export SubmitActionResult
func SubmitActionResult(appId string, callId string, result string) {
	// TODO
}
