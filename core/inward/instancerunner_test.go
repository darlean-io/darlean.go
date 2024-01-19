package inward

import (
	"fmt"
	"testing"
	"time"

	"github.com/darlean-io/darlean.go/base/actionerror"
	"github.com/darlean-io/darlean.go/core/wire"
	"github.com/darlean-io/darlean.go/utils/checks"
	. "github.com/darlean-io/darlean.go/utils/variant"
)

func TestInstanceRunner_Exclusive(t *testing.T) {
	runner, wrapper := newRunner()

	var results []string

	handleResult := func(result any, err *actionerror.Error) {
		if err != nil {
			results = append(results, fmt.Sprintf("ERR:%v", err.Code))
		} else {
			results = append(results, fmt.Sprintf("%v", result))
		}
	}

	runner.Invoke(&wire.ActorCallRequestIn{ActionName: "Exclusive", Arguments: []Assignable{FromString("Hello")}}, handleResult)
	runner.Invoke(&wire.ActorCallRequestIn{ActionName: "Exclusive", Arguments: []Assignable{FromString("World")}}, handleResult)

	time.Sleep(SLEEP_BASIS * 3)

	runner.TriggerDeactivate()
	time.Sleep(time.Second)

	runner.Invoke(&wire.ActorCallRequestIn{ActionName: "Exclusive", Arguments: []Assignable{FromString("Too late")}}, handleResult)

	time.Sleep(time.Second)

	checks.Equal(t, []string{
		"Activate",
		"Activated",
		"Perform {exclusive} with {Hello}",
		"Performed {exclusive} with {Hello}",
		"Perform {exclusive} with {World}",
		"Performed {exclusive} with {World}",
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

	handleResult := func(result any, err *actionerror.Error) {
		if err != nil {
			results = append(results, fmt.Sprintf("ERR:%v", err.Code))
		} else {
			results = append(results, fmt.Sprintf("%v", result))
		}
	}

	runner.Invoke(&wire.ActorCallRequestIn{ActionName: "Shared", Arguments: []Assignable{FromString("Hello")}}, handleResult)
	time.Sleep(SLEEP_BASIS_HALF)
	runner.Invoke(&wire.ActorCallRequestIn{ActionName: "Shared", Arguments: []Assignable{FromString("World")}}, handleResult)

	time.Sleep(SLEEP_BASIS * 2)

	runner.Invoke(&wire.ActorCallRequestIn{ActionName: "Shared", Arguments: []Assignable{FromString("Hello2")}}, handleResult)
	time.Sleep(SLEEP_BASIS_HALF)
	runner.Invoke(&wire.ActorCallRequestIn{ActionName: "Shared", Arguments: []Assignable{FromString("World2")}}, handleResult)
	time.Sleep(time.Second)

	runner.TriggerDeactivate()
	time.Sleep(SLEEP_BASIS_HALF)
	runner.Invoke(&wire.ActorCallRequestIn{ActionName: "Shared", Arguments: []Assignable{FromString("Too late")}}, handleResult)
	time.Sleep(time.Second)
	runner.Invoke(&wire.ActorCallRequestIn{ActionName: "Shared", Arguments: []Assignable{FromString("Too late")}}, handleResult)

	time.Sleep(time.Second)

	checks.Equal(t, []string{
		"Activate",
		"Activated",
		"Perform {shared} with {Hello}",
		"Perform {shared} with {World}",
		"Performed {shared} with {Hello}",
		"Performed {shared} with {World}",
		"Perform {shared} with {Hello2}",
		"Perform {shared} with {World2}",
		"Performed {shared} with {Hello2}",
		"Performed {shared} with {World2}",
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

	handleResult := func(result any, err *actionerror.Error) {
		if err != nil {
			results = append(results, fmt.Sprintf("ERR:%v", err.Code))
		} else {
			results = append(results, fmt.Sprintf("%v", result))
		}
	}

	// Use a slightly faster "none" fuction (with shorter delay than the "activate" that runs in parallel)
	// to avoid race conditions when the activate and none return.
	runner.Invoke(&wire.ActorCallRequestIn{ActionName: "NoneFaster", Arguments: []Assignable{FromString("Hello")}}, handleResult)
	time.Sleep(SLEEP_BASIS_HALF)
	runner.Invoke(&wire.ActorCallRequestIn{ActionName: "None", Arguments: []Assignable{FromString("World")}}, handleResult)
	time.Sleep(SLEEP_BASIS)
	runner.Invoke(&wire.ActorCallRequestIn{ActionName: "None", Arguments: []Assignable{FromString("Foo")}}, handleResult)

	time.Sleep(time.Second)

	runner.TriggerDeactivate()
	time.Sleep(SLEEP_BASIS_HALF)
	runner.Invoke(&wire.ActorCallRequestIn{ActionName: "None", Arguments: []Assignable{FromString("During-deactivate")}}, handleResult)
	time.Sleep(SLEEP_BASIS)
	runner.Invoke(&wire.ActorCallRequestIn{ActionName: "None", Arguments: []Assignable{FromString("Too late")}}, handleResult)

	time.Sleep(time.Second)

	// The order of first 2 items can vary depending on thread scheduling. So create
	// two truth items for these cases.
	truth := []any{[]string{
		"Activate",
		"Perform {nonefaster} with {Hello}",
		"Perform {none} with {World}",
		"Performed {nonefaster} with {Hello}",
		"Activated",
		"Performed {none} with {World}",
		"Perform {none} with {Foo}",
		"Performed {none} with {Foo}",
		"Deactivate",
		"Perform {none} with {During-deactivate}",
		"Deactivated",
		"Performed {none} with {During-deactivate}",
	}, []string{
		"Perform {nonefaster} with {Hello}",
		"Activate",
		"Perform {none} with {World}",
		"Performed {nonefaster} with {Hello}",
		"Activated",
		"Performed {none} with {World}",
		"Perform {none} with {Foo}",
		"Performed {none} with {Foo}",
		"Deactivate",
		"Perform {none} with {During-deactivate}",
		"Deactivated",
		"Performed {none} with {During-deactivate}",
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
