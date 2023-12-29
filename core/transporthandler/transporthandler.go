package transporthandler

import (
	"fmt"
	"sync"

	"github.com/darlean-io/darlean.go/base/invoker"
	"github.com/darlean-io/darlean.go/core"
	"github.com/darlean-io/darlean.go/core/invoke"
	"github.com/darlean-io/darlean.go/core/wire"

	"github.com/google/uuid"
)

type pendingCall struct {
	finished chan<- *invoker.Response
}

type TransportHandler struct {
	appId        string
	transport    core.Transport
	pendingCalls map[string]pendingCall
	mutex        sync.Mutex
	dispatcher   InwardCallDispatcher
}

type InwardCallDispatcher interface {
	Dispatch(tags *wire.ActorCallRequest, onFinished func(*wire.ActorCallResponse))
}

func (invoker *TransportHandler) Listen() {
	for tags := range invoker.transport.GetInputChannel() {
		switch tags.Remotecall_Kind {
		case "call":
			if invoker.dispatcher == nil {
				fmt.Println("transporthandler: Ignore incoming message: no dispatcher assigned")
				continue
			}

			go func() {
				invoker.dispatcher.Dispatch(&tags.ActorCallRequest, func(response *wire.ActorCallResponse) {
					responseMsg := wire.Tags{
						TransportTags: wire.TransportTags{
							Transport_Receiver: tags.Transport_Return,
							Transport_Return:   invoker.appId,
						},
						RemoteCallTags: wire.RemoteCallTags{
							Remotecall_Kind: "return",
							Remotecall_Id:   tags.Remotecall_Id,
						},
						ActorCallResponse: *response,
					}
					invoker.transport.Send(responseMsg)
				})
			}()

			continue
		case "return":
			invoker.handleReturnMessage(tags)
		}
	}
}

func (handler *TransportHandler) handleReturnMessage(tags *wire.Tags) {
	handler.mutex.Lock()
	call, found := handler.pendingCalls[tags.Remotecall_Id]
	if found {
		delete(handler.pendingCalls, tags.Remotecall_Id)
	}
	handler.mutex.Unlock()

	if !found {
		fmt.Println("Received value without matching call")
		return
	}

	call.finished <- &invoker.Response{
		Value: tags.Value,
		Error: tags.Error,
	}
}

func New(transport core.Transport, dispatcher InwardCallDispatcher, appId string) *TransportHandler {
	invoker := TransportHandler{
		appId:        appId,
		transport:    transport,
		dispatcher:   dispatcher,
		pendingCalls: make(map[string]pendingCall),
	}

	go invoker.Listen()

	return &invoker
}

func (handler *TransportHandler) Invoke(req *invoke.TransportHandlerInvokeRequest) *invoker.Response {
	id := uuid.NewString()

	tags := wire.Tags{}
	tags.Transport_Receiver = req.Receiver
	tags.Transport_Return = handler.appId
	tags.Remotecall_Id = id
	tags.Remotecall_Kind = "call"
	tags.ActorType = req.ActorType
	tags.ActorId = req.ActorId
	tags.ActionName = req.ActionName
	tags.Arguments = req.Parameters

	response := make(chan *invoker.Response)

	handler.mutex.Lock()
	handler.pendingCalls[id] = pendingCall{
		finished: response,
	}
	handler.mutex.Unlock()

	err := handler.transport.Send(tags)
	if err != nil {
		panic(err)
	}

	return <-response
}
