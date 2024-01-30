package main

type RegisterActorOptions struct {
	ActorType string
}

type RegisterActorResult struct {
	ActorHandle string
}

type RegisterActionOptions struct {
	ActorType  string
	ActionName string
	Locking    string
}

type SubmitActionResultOptions struct {
	Value any
	Error *ActionError
}

type PerformActionOptions struct {
	ActorType  string
	ActorId    []string
	ActionName string
	Arguments  []any
}

type InvokeActionOptions struct {
	ActorType  string
	ActorId    []string
	ActionName string
	Arguments  []any
}

type InvokeActionResult struct {
	Value any
	Error *ActionError
}

type ActionErrorKind string

const ERROR_KIND_FRAMEWORK = ActionErrorKind("framework")
const ERROR_KIND_APPLICATION = ActionErrorKind("application")

type ActionError struct {
	Code       string          `json:"code"`
	Message    string          `json:"message"`
	Template   string          `json:"template"`
	Kind       ActionErrorKind `json:"kind"`
	Parameters map[string]any  `json:"parameters"`
	Nested     []*ActionError  `json:"nested"`
	Stack      string          `json:"stack"`
}

// Satisfies the error type.
func (error *ActionError) Error() string {
	return error.Message
}
