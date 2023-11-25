package invoke

import (
	"core"
	"core/wire"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type pendingCall struct {
	finished chan<- *InvokeResponse
}

type StaticInvoker struct {
	appId        string
	transport    core.Transport
	pendingCalls map[string]pendingCall
	mutex        sync.Mutex
}

type StaticInvokeRequest struct {
	InvokeRequest
	Receiver string
}

func (invoker *StaticInvoker) Listen() {
	for tags := range invoker.transport.GetInputChannel() {
		switch tags.Remotecall_Kind {
		case "call":
			fmt.Println("Ignore incoming message: 'call' is not yet implemented")
			continue
		case "return":
			invoker.handleReturnMessage(tags)
		}
	}
}

func (invoker *StaticInvoker) handleReturnMessage(tags *wire.Tags) {
	invoker.mutex.Lock()
	call, found := invoker.pendingCalls[tags.Remotecall_Id]
	if found {
		delete(invoker.pendingCalls, tags.Remotecall_Id)
	}
	invoker.mutex.Unlock()

	if !found {
		fmt.Println("Received value without matching call")
		return
	}

	call.finished <- &InvokeResponse{
		Value: tags.Value,
		Error: tags.Error,
	}
}

func NewStaticInvoker(transport core.Transport, appId string) *StaticInvoker {
	invoker := StaticInvoker{
		appId:        appId,
		transport:    transport,
		pendingCalls: make(map[string]pendingCall),
	}

	go invoker.Listen()

	return &invoker
}

func (invoker *StaticInvoker) Invoke(req *StaticInvokeRequest) *InvokeResponse {
	id := uuid.NewString()

	tags := wire.Tags{}
	tags.Transport_Receiver = req.Receiver
	tags.Transport_Return = invoker.appId
	tags.Remotecall_Id = id
	tags.Remotecall_Kind = "call"
	tags.ActorType = req.ActorType
	tags.ActorId = req.ActorId
	tags.ActionName = req.ActionName
	tags.Arguments = req.Parameters

	response := make(chan *InvokeResponse)

	invoker.mutex.Lock()
	invoker.pendingCalls[id] = pendingCall{
		finished: response,
	}
	invoker.mutex.Unlock()

	err := invoker.transport.Send(tags)
	if err != nil {
		panic(err)
	}

	return <-response
}
