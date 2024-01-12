/*
Package actionerror defines the types and functions for structured handling of action errors within a Darlean cluster.
*/
package actionerror

import (
	"fmt"
	"strings"

	"github.com/darlean-io/darlean.go/utils/variant"
)

type Kind string

const ERROR_KIND_FRAMEWORK = Kind("framework")
const ERROR_KIND_APPLICATION = Kind("application")

/*
Error represents an error that occurred while invoking an action.

Action errors can either be application errors (that occurred in application logic) or
framework errors (that occurred within Darlean, like for network errors).

Action errors are serializable and can passed to and understood by remote actors, even in different languages.
*/
type Error struct {
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	Template   string         `json:"template"`
	Kind       Kind           `json:"kind"`
	Parameters map[string]any `json:"parameters"`
	Nested     []*Error       `json:"nested"`
	Stack      string         `json:"stack"`
}

/*
Options can be used to create a new [base.ActionError] via [base.NewFrameworkError]
or [base.NewApplicationError].
*/
type Options struct {
	Code       string
	Template   string
	Parameters map[string]any
	Nested     []*Error
	Stack      string
}

// Satisfies the error type.
func (error Error) Error() string {
	return error.Message
}

/*
Formats a template string by replacing placeholders with the specified parameters.

Placeholders must be of the form `[Name]`, where Name should be one of the keys
of the parameters.
*/
func FormatTemplate(template string, parameters map[string]any) string {
	message := template
	for key, value := range parameters {
		assignable, supported := value.(variant.Assignable)
		if supported {
			var target any
			err := assignable.AssignTo(target)
			if err != nil {
				value = err.Error()
			} else {
				value = target
			}
		}
		message = strings.ReplaceAll(message, "["+key+"]", "\""+fmt.Sprint(value)+"\"")
	}
	return message
}

func newActionError(options Options, kind Kind) *Error {
	e := Error{
		Kind:       kind,
		Code:       options.Code,
		Message:    options.Template,
		Template:   options.Template,
		Parameters: options.Parameters,
		Nested:     options.Nested,
		Stack:      options.Stack,
	}
	if len(e.Parameters) > 0 {
		e.Message = FormatTemplate(options.Template, e.Parameters)
	}
	if e.Code != "" {
		if e.Message == "" {
			e.Message = e.Code
		} else {
			e.Message = "(" + e.Code + ") " + e.Message
		}
	}
	return &e
}

// Returns a new framework action error for the specified options. Application code
// should not use this function, but should use [base.NewApplicationError] instead.
func NewFrameworkError(options Options) *Error {
	return newActionError(options, ERROR_KIND_FRAMEWORK)
}

// Returns a new application action error for the specified options. Application code
// is encouraged to use this mechanism instead of regular errors, because it provides a
// consistent way of propagating errors throughout the Darlean cluster.
func NewApplicationError(options Options) *Error {
	return newActionError(options, ERROR_KIND_APPLICATION)
}
