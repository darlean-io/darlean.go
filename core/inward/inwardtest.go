package inward

import (
	"core/normalized"
	"fmt"
	"strings"
	"time"
)

type TestActorWrapper struct {
	history []string
	id      string
}

const SLEEP_BASIS_TENTH = time.Millisecond * 10

const SLEEP_BASIS_HALF = SLEEP_BASIS_TENTH * 5
const SLEEP_BASIS_SHORT = SLEEP_BASIS_TENTH * 8
const SLEEP_BASIS = SLEEP_BASIS_TENTH * 10

func (wrapper *TestActorWrapper) Activate() error {
	wrapper.history = append(wrapper.history, "Activate")
	time.Sleep(SLEEP_BASIS)
	wrapper.history = append(wrapper.history, "Activated")
	return nil
}

func (wrapper *TestActorWrapper) Deactivate() error {
	wrapper.history = append(wrapper.history, "Deactivate")
	time.Sleep(SLEEP_BASIS)
	wrapper.history = append(wrapper.history, "Deactivated")
	return nil
}

func (wrapper *TestActorWrapper) Perform(actionName normalized.ActionName, args []any) (result any, err error) {
	wrapper.history = append(wrapper.history, fmt.Sprintf("Perform {%v} with {%v}", string(actionName), args[0]))
	if strings.Contains(string(actionName), "faster") {
		time.Sleep(SLEEP_BASIS_SHORT)
	} else {
		time.Sleep(SLEEP_BASIS)
	}
	wrapper.history = append(wrapper.history, fmt.Sprintf("Performed {%v} with {%v}", string(actionName), args[0]))
	resultstring := strings.ToLower(args[0].(string))
	if wrapper.id != "" {
		resultstring = wrapper.id + ":" + resultstring
	}
	return resultstring, nil
}

func GetTestActionDefs() map[normalized.ActionName]ActionDef {
	return map[normalized.ActionName]ActionDef{
		"exclusive":  {locking: ACTION_LOCK_EXCLUSIVE},
		"shared":     {locking: ACTION_LOCK_SHARED},
		"none":       {locking: ACTION_LOCK_NONE},
		"nonefaster": {locking: ACTION_LOCK_NONE},
	}
}

func newRunner() (*DefaultInstanceRunner, *TestActorWrapper) {
	var wrapper TestActorWrapper

	runner := NewInstanceRunner(&wrapper, false, GetTestActionDefs(), nil)
	return runner, &wrapper
}
