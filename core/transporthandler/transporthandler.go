package transporthandler

import (
	"fmt"
	"sync"

	"github.com/darlean-io/darlean.go/base/invoker"
	"github.com/darlean-io/darlean.go/core"
	"github.com/darlean-io/darlean.go/core/invoke"
	"github.com/darlean-io/darlean.go/core/inward"
	"github.com/darlean-io/darlean.go/core/wire"

	"github.com/google/uuid"
)

type pendingCall struct {
	finished chan<- *invoker.Response
}

/*
TransportHandler handles the incoming and outgoing calls to a transport. Implements [TransportInvoker].
*/
type TransportHandler struct {
	appId             string
	transport         core.Transport
	pendingCalls      map[string]pendingCall
	mutex             sync.Mutex
	dispatcherFactory func() InwardCallDispatcher
	dispatcher        InwardCallDispatcher
}

type InwardCallDispatcher interface {
	Dispatch(tags *wire.ActorCallRequestIn, onFinished func(*wire.ActorCallResponseOut))
}

func (invoker *TransportHandler) Listen() {
	if invoker.dispatcherFactory != nil {
		invoker.dispatcher = invoker.dispatcherFactory()
	}

	for tags := range invoker.transport.GetInputChannel() {
		switch tags.Remotecall_Kind {
		case "call":
			if invoker.dispatcher == nil {
				fmt.Println("transporthandler: Ignore incoming message: no dispatcher assigned")
				continue
			}

			go func(tags *wire.TagsIn) {
				invoker.dispatcher.Dispatch(&tags.ActorCallRequestIn, func(response *wire.ActorCallResponseOut) {
					responseMsg := wire.TagsOut{
						TransportTags: wire.TransportTags{
							Transport_Receiver: tags.Transport_Return,
							Transport_Return:   invoker.appId,
						},
						RemoteCallTags: wire.RemoteCallTags{
							Remotecall_Kind: "return",
							Remotecall_Id:   tags.Remotecall_Id,
						},
						ActorCallResponseOut: *response,
					}
					invoker.transport.Send(responseMsg)
				})
			}(tags)

			continue
		case "return":
			invoker.handleReturnMessage(tags)
		}
	}
}

func (handler *TransportHandler) handleReturnMessage(tags *wire.TagsIn) {
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

	// fmt.Printf("TransportHandler received %+v\n", tags)

	call.finished <- &invoker.Response{
		Value: tags.Value,
		Error: tags.Error,
	}
}

func New(transport core.Transport, appId string) *TransportHandler {
	invoker := TransportHandler{
		appId:        appId,
		transport:    transport,
		pendingCalls: make(map[string]pendingCall),
	}

	return &invoker
}

func (handler *TransportHandler) Start(dispatcher *inward.Dispatcher) {
	handler.dispatcher = dispatcher
	go handler.Listen()
}

// Invoke invokes a remote action and satisfies [TransportInvoker.Invoke]
func (handler *TransportHandler) Invoke(req *invoke.TransportHandlerInvokeRequest) *invoker.Response {
	id := uuid.NewString()

	tags := wire.TagsOut{}
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
		fmt.Printf("Nats error to %s for %s.%s: %+v\n", req.Receiver, req.ActorType, req.ActionName, err)
		panic(err)
	}

	return <-response
}
