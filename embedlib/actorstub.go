package main

import (
	"fmt"

	"github.com/darlean-io/darlean.go/core/normalized"
	"github.com/darlean-io/darlean.go/utils/variant"
)

type ActorStub struct {
	id   []string
	info *ActorInfo
}

func (a *ActorStub) Activate() error {
	fmt.Printf("Activating embed actor. TODO: Via callback.\n")
	return nil
}

func (a *ActorStub) Deactivate() error {
	fmt.Printf("Deactivating embed actor. TODO: Via callback.\n")
	return nil
}

func (a *ActorStub) Perform(actionName normalized.ActionName, args []variant.Assignable) (result any, err error) {
	action := a.info.Actions[actionName]

	channel := make(chan SubmitActionResultOptions)

	a.info.CallManager.MakeActionCall(&action, channel, args)
	callResult := <-channel

	return callResult.Value, makeTrulyNil(callResult.Error)
}

func NewActorStub(info *ActorInfo, id []string) *ActorStub {
	return &ActorStub{
		info: info,
		id:   id,
	}
}

func makeTrulyNil(input *ActionError) error {
	// Danger: In go, nil is not always nil. If input is nil, it is a (ActionError)(nil), which
	// is different from a (any)(nil) that other code that just expects an error will be using to
	// check for nil. Therefore, we check for (ActionError)(nil) (simply by comparing with nil,
	// which go internally converts to (ActionError)(nil) because it knows the type of input)
	// and return a "regular" nil. See https://codefibershq.com/blog/golang-why-nil-is-not-always-nil
	var e error
	if input != nil {
		e = input
	}
	return e
}
