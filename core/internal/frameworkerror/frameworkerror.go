package frameworkerror

import "github.com/darlean-io/darlean.go/base/actionerror"

// Returns a new framework action error for the specified options. Application code
// should not use this function, but should use [base.NewApplicationError] instead.
func New(options actionerror.Options) *actionerror.Error {
	return actionerror.NewActionError(options, actionerror.ERROR_KIND_FRAMEWORK)
}

func FromError(e error) *actionerror.Error {
	if e == nil {
		return nil
	}
	return New(actionerror.Options{
		Template: e.Error(),
	})
}
