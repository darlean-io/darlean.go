package invoke

import "github.com/darlean-io/darlean.go/base/invoker"

const FRAMEWORK_ERROR_PARAMETER_REDIRECT_DESTINATION = "REDIRECT_DESTINATION"
const FRAMEWORK_ERROR_INVOKE_ERROR = "INVOKE_ERROR"
const FRAMEWORK_ERROR_NO_RECEIVERS_AVAILABLE = "NO_RECEIVERS_AVAILABLE"

type TransportHandlerInvokeRequest struct {
	invoker.Request
	Receiver string
}

type TransportInvoker interface {
	Invoke(req *TransportHandlerInvokeRequest) *invoker.Response
}
