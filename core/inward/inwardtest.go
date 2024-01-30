package inward

import (
	"fmt"
	"strings"
	"time"

	"github.com/darlean-io/darlean.go/base/actionerror"
	"github.com/darlean-io/darlean.go/core/normalized"
	"github.com/darlean-io/darlean.go/utils/variant"
)

type TestActorWrapper struct {
	history []string
	id      string
}

const SLEEP_BASIS_TENTH = time.Millisecond * 10

const SLEEP_BASIS_HALF = SLEEP_BASIS_TENTH * 5
const SLEEP_BASIS_SHORT = SLEEP_BASIS_TENTH * 8
const SLEEP_BASIS = SLEEP_BASIS_TENTH * 10

func (wrapper *TestActorWrapper) Create() *actionerror.Error {
	wrapper.history = append(wrapper.history, "Create")
	time.Sleep(SLEEP_BASIS)
	wrapper.history = append(wrapper.history, "Created")
	return nil
}

func (wrapper *TestActorWrapper) Activate() *actionerror.Error {
	wrapper.history = append(wrapper.history, "Activate")
	time.Sleep(SLEEP_BASIS)
	wrapper.history = append(wrapper.history, "Activated")
	return nil
}

func (wrapper *TestActorWrapper) Deactivate() *actionerror.Error {
	wrapper.history = append(wrapper.history, "Deactivate")
	time.Sleep(SLEEP_BASIS)
	wrapper.history = append(wrapper.history, "Deactivated")
	return nil
}

func (wrapper *TestActorWrapper) Release() *actionerror.Error {
	wrapper.history = append(wrapper.history, "Release")
	time.Sleep(SLEEP_BASIS)
	wrapper.history = append(wrapper.history, "Released")
	return nil
}

func (wrapper *TestActorWrapper) Perform(actionName normalized.ActionName, args []variant.Assignable) (result any, err *actionerror.Error) {
	wrapper.history = append(wrapper.history, fmt.Sprintf("Perform {%v} with {%v}", string(actionName), args[0]))
	if strings.Contains(string(actionName), "faster") {
		time.Sleep(SLEEP_BASIS_SHORT)
	} else {
		time.Sleep(SLEEP_BASIS)
	}
	wrapper.history = append(wrapper.history, fmt.Sprintf("Performed {%v} with {%v}", string(actionName), args[0]))
	arg0, err0 := args[0].AssignToString()
	resultstring := strings.ToLower(arg0)
	if wrapper.id != "" {
		resultstring = wrapper.id + ":" + resultstring
	}
	return resultstring, actionerror.FromError(err0)
}

func GetTestActionDefs() map[normalized.ActionName]ActionDef {
	return map[normalized.ActionName]ActionDef{
		"exclusive":  {Locking: ACTION_LOCK_EXCLUSIVE},
		"shared":     {Locking: ACTION_LOCK_SHARED},
		"none":       {Locking: ACTION_LOCK_NONE},
		"nonefaster": {Locking: ACTION_LOCK_NONE},
	}
}

func newRunner() (*DefaultInstanceRunner, *TestActorWrapper) {
	var wrapper TestActorWrapper

	runner := NewInstanceRunner(&wrapper, normalized.NormalizeActorType("TestActor"), []string{"123"}, false, GetTestActionDefs(), nil)
	return runner, &wrapper
}
