package invoke

const FRAMEWORK_ERROR_PARAMETER_REDIRECT_DESTINATION = "REDIRECT_DESTINATION"
const FRAMEWORK_ERROR_INVOKE_ERROR = "INVOKE_ERROR"
const FRAMEWORK_ERROR_NO_RECEIVERS_AVAILABLE = "NO_RECEIVERS_AVAILABLE"

type InvokeRequest struct {
	ActorType  string
	ActorId    []string
	ActionName string
	Parameters []any
	Lazy       bool
}

type InvokeResponse struct {
	Error any
	Value any
}
