package inward

import (
	"core/checks"
	"core/normalized"
	"core/wire"
	"fmt"
	"strings"
	"testing"
	"time"
)

type TestWrapper struct {
	history []string
}

const SLEEP_BASIS_TENTH = time.Millisecond * 10

const SLEEP_BASIS_HALF = SLEEP_BASIS_TENTH * 5
const SLEEP_BASIS_SHORT = SLEEP_BASIS_TENTH * 8
const SLEEP_BASIS = SLEEP_BASIS_TENTH * 10

func (wrapper *TestWrapper) Activate() error {
	wrapper.history = append(wrapper.history, "Activate")
	time.Sleep(SLEEP_BASIS)
	wrapper.history = append(wrapper.history, "Activated")
	return nil
}

func (wrapper *TestWrapper) Deactivate() error {
	wrapper.history = append(wrapper.history, "Deactivate")
	time.Sleep(SLEEP_BASIS)
	wrapper.history = append(wrapper.history, "Deactivated")
	return nil
}

func (wrapper *TestWrapper) Perform(actionName normalized.ActionName, args []any) (result any, err error) {
	wrapper.history = append(wrapper.history, fmt.Sprintf("Perform {%v} with {%v}", string(actionName), args[0]))
	if actionName == "NoneFaster" {
		time.Sleep(SLEEP_BASIS_SHORT)
	} else {
		time.Sleep(SLEEP_BASIS)
	}
	wrapper.history = append(wrapper.history, fmt.Sprintf("Performed {%v} with {%v}", string(actionName), args[0]))
	resultstring := strings.ToLower(args[0].(string))
	return resultstring, nil
}

func newRunner() (*DefaultInstanceRunner, *TestWrapper) {
	actionDefs := map[normalized.ActionName]ActionDef{
		"exclusive":  {locking: ACTION_LOCK_EXCLUSIVE},
		"shared":     {locking: ACTION_LOCK_SHARED},
		"none":       {locking: ACTION_LOCK_NONE},
		"nonefaster": {locking: ACTION_LOCK_NONE},
	}

	var wrapper TestWrapper

	runner := NewInstanceRunner(&wrapper, false, actionDefs, nil)
	return runner, &wrapper
}

func TestInstanceRunner_Exclusive(t *testing.T) {
	runner, wrapper := newRunner()

	var results []string

	handleResult := func(result any, err error) {
		if err != nil {
			results = append(results, fmt.Sprintf("ERR:%v", err))
		} else {
			results = append(results, fmt.Sprintf("%v", result))
		}
	}

	runner.Invoke(&wire.ActorCallRequest{ActionName: "Exclusive", Arguments: []any{"Hello"}}, handleResult)
	runner.Invoke(&wire.ActorCallRequest{ActionName: "Exclusive", Arguments: []any{"World"}}, handleResult)

	time.Sleep(SLEEP_BASIS * 3)

	runner.TriggerDeactivate()
	time.Sleep(time.Second)

	runner.Invoke(&wire.ActorCallRequest{ActionName: "Exclusive", Arguments: []any{"Too late"}}, handleResult)

	time.Sleep(time.Second)

	checks.Equal(t, []string{
		"Activate",
		"Activated",
		"Perform {Exclusive} with {Hello}",
		"Performed {Exclusive} with {Hello}",
		"Perform {Exclusive} with {World}",
		"Performed {Exclusive} with {World}",
		"Deactivate",
		"Deactivated",
	}, wrapper.history, "Events should be as expected")

	checks.Equal(t, []string{
		"hello",
		"world",
		"ERR:DEACTIVATED",
	}, results, "Results should be as expected")
}

func TestInstanceRunner_Shared(t *testing.T) {
	runner, wrapper := newRunner()
	var results []string

	handleResult := func(result any, err error) {
		if err != nil {
			results = append(results, fmt.Sprintf("ERR:%v", err))
		} else {
			results = append(results, fmt.Sprintf("%v", result))
		}
	}

	runner.Invoke(&wire.ActorCallRequest{ActionName: "Shared", Arguments: []any{"Hello"}}, handleResult)
	time.Sleep(SLEEP_BASIS_HALF)
	runner.Invoke(&wire.ActorCallRequest{ActionName: "Shared", Arguments: []any{"World"}}, handleResult)

	time.Sleep(SLEEP_BASIS * 2)

	runner.Invoke(&wire.ActorCallRequest{ActionName: "Shared", Arguments: []any{"Hello2"}}, handleResult)
	time.Sleep(SLEEP_BASIS_HALF)
	runner.Invoke(&wire.ActorCallRequest{ActionName: "Shared", Arguments: []any{"World2"}}, handleResult)
	time.Sleep(time.Second)

	runner.TriggerDeactivate()
	time.Sleep(SLEEP_BASIS_HALF)
	runner.Invoke(&wire.ActorCallRequest{ActionName: "Shared", Arguments: []any{"Too late"}}, handleResult)
	time.Sleep(time.Second)
	runner.Invoke(&wire.ActorCallRequest{ActionName: "Shared", Arguments: []any{"Too late"}}, handleResult)

	time.Sleep(time.Second)

	checks.Equal(t, []string{
		"Activate",
		"Activated",
		"Perform {Shared} with {Hello}",
		"Perform {Shared} with {World}",
		"Performed {Shared} with {Hello}",
		"Performed {Shared} with {World}",
		"Perform {Shared} with {Hello2}",
		"Perform {Shared} with {World2}",
		"Performed {Shared} with {Hello2}",
		"Performed {Shared} with {World2}",
		"Deactivate",
		"Deactivated",
	}, wrapper.history, "Events should be as expected")

	checks.Equal(t, []string{
		"hello",
		"world",
		"hello2",
		"world2",
		"ERR:DEACTIVATED",
		"ERR:DEACTIVATED",
	}, results, "Results should be as expected")
}

func TestInstanceRunner_None(t *testing.T) {
	runner, wrapper := newRunner()
	var results []string

	handleResult := func(result any, err error) {
		if err != nil {
			results = append(results, fmt.Sprintf("ERR:%v", err))
		} else {
			results = append(results, fmt.Sprintf("%v", result))
		}
	}

	// Use a slightly faster "none" fuction (with shorter delay than the "activate" that runs in parallel)
	// to avoid race conditions when the activate and none return.
	runner.Invoke(&wire.ActorCallRequest{ActionName: "NoneFaster", Arguments: []any{"Hello"}}, handleResult)
	time.Sleep(SLEEP_BASIS_HALF)
	runner.Invoke(&wire.ActorCallRequest{ActionName: "None", Arguments: []any{"World"}}, handleResult)
	time.Sleep(SLEEP_BASIS)
	runner.Invoke(&wire.ActorCallRequest{ActionName: "None", Arguments: []any{"Foo"}}, handleResult)

	time.Sleep(time.Second)

	runner.TriggerDeactivate()
	time.Sleep(SLEEP_BASIS_HALF)
	runner.Invoke(&wire.ActorCallRequest{ActionName: "None", Arguments: []any{"During-deactivate"}}, handleResult)
	time.Sleep(SLEEP_BASIS)
	runner.Invoke(&wire.ActorCallRequest{ActionName: "None", Arguments: []any{"Too late"}}, handleResult)

	time.Sleep(time.Second)

	// The order of first 2 items can vary depending on thread scheduling. So create
	// two truth items for these cases.
	truth := []any{[]string{
		"Activate",
		"Perform {NoneFaster} with {Hello}",
		"Perform {None} with {World}",
		"Performed {NoneFaster} with {Hello}",
		"Activated",
		"Performed {None} with {World}",
		"Perform {None} with {Foo}",
		"Performed {None} with {Foo}",
		"Deactivate",
		"Perform {None} with {During-deactivate}",
		"Deactivated",
		"Performed {None} with {During-deactivate}",
	}, []string{
		"Perform {NoneFaster} with {Hello}",
		"Activate",
		"Perform {None} with {World}",
		"Performed {NoneFaster} with {Hello}",
		"Activated",
		"Performed {None} with {World}",
		"Perform {None} with {Foo}",
		"Performed {None} with {Foo}",
		"Deactivate",
		"Perform {None} with {During-deactivate}",
		"Deactivated",
		"Performed {None} with {During-deactivate}",
	}}
	checks.EqualOneOf(t, truth, wrapper.history, "Events should be as expected")

	checks.Equal(t, []string{
		"hello",
		"world",
		"foo",
		"during-deactivate",
		"ERR:DEACTIVATED",
	}, results, "Results should be as expected")
}
