package invoke

import (
	"core/variant"
	"fmt"
	"strings"
)

const FRAMEWORK_ERROR_PARAMETER_REDIRECT_DESTINATION = "REDIRECT_DESTINATION"
const FRAMEWORK_ERROR_INVOKE_ERROR = "INVOKE_ERROR"

type InvokeRequest struct {
	ActorType  string
	ActorId    []string
	ActionName string
	Parameters []variant.Variant
	Lazy       bool
}

type InvokeResponse struct {
	Error variant.Variant
	Value variant.Variant
}

type ErrorKind string

const ERROR_KIND_FRAMEWORK = "framework"
const ERROR_KIND_APPLICATION = "application"

type ActionError struct {
	Code       string                     `json:"code"`
	Message    string                     `json:"message"`
	Template   string                     `json:"template"`
	Kind       ErrorKind                  `json:"kind"`
	Parameters map[string]variant.Variant `json:"parameters"`
	Nested     []*ActionError             `json:"nested"`
	Stack      string                     `json:"stack"`
}

type ActionErrorOptions struct {
	Code       string
	Template   string
	Parameters map[string]any
	Nested     []*ActionError
	Stack      string
}

func (error ActionError) Error() string {
	return error.Message
}

func FormatTemplate(template string, parameters map[string]variant.Variant) string {
	message := template
	for key, value := range parameters {
		v, ok := value.GetDirect()
		if !ok {
			value.Get(v)
		}
		message = strings.ReplaceAll(message, "["+key+"]", "\""+fmt.Sprint(v)+"\"")
	}
	return message
}

func newActionError(options ActionErrorOptions, kind ErrorKind) *ActionError {
	e := ActionError{
		Kind:       kind,
		Code:       options.Code,
		Message:    options.Template,
		Template:   options.Template,
		Parameters: variant.Map(options.Parameters),
		Nested:     options.Nested,
		Stack:      options.Stack,
	}
	if len(e.Parameters) > 0 {
		e.Message = FormatTemplate(options.Template, e.Parameters)
	}
	return &e
}

func NewFrameworkError(options ActionErrorOptions) *ActionError {
	return newActionError(options, ERROR_KIND_FRAMEWORK)
}

func NewApplicationError(options ActionErrorOptions) *ActionError {
	return newActionError(options, ERROR_KIND_APPLICATION)
}
