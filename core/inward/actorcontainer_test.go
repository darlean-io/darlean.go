package inward

import (
	"fmt"
	"testing"
	"time"

	"github.com/darlean-io/darlean.go/base/actionerror"
	"github.com/darlean-io/darlean.go/core/normalized"
	"github.com/darlean-io/darlean.go/core/wire"

	"github.com/darlean-io/darlean.go/utils/checks"
	. "github.com/darlean-io/darlean.go/utils/variant"
)

func TestActorContainer(t *testing.T) {
	wrapperFactory := func(id []string) InstanceWrapper {
		wrapper := TestActorWrapper{
			id: id[0],
		}
		return &wrapper
	}

	var results []string

	container := NewStandardActorContainer(normalized.NormalizeActorType("TestActor"), false, GetTestActionDefs(), wrapperFactory, func() {
		results = append(results, "CONTAINER-STOPPED")
	})

	handleResult := func(result any, err *actionerror.Error) {
		if err != nil {
			results = append(results, fmt.Sprintf("ERR:%v", err.Code))
		} else {
			results = append(results, fmt.Sprintf("%v", result))
		}
	}

	container.Dispatch(&wire.ActorCallRequestIn{ActorId: []string{"123"}, ActionName: "Exclusive", Arguments: []Assignable{FromString("Hello")}}, handleResult)
	container.Dispatch(&wire.ActorCallRequestIn{ActorId: []string{"123"}, ActionName: "Exclusive", Arguments: []Assignable{FromString("World")}}, handleResult)
	container.Dispatch(&wire.ActorCallRequestIn{ActorId: []string{"234"}, ActionName: "Exclusive", Arguments: []Assignable{FromString("Moon")}}, handleResult)

	time.Sleep(SLEEP_BASIS * 5)

	go container.Stop()

	time.Sleep(SLEEP_BASIS_HALF)

	container.Dispatch(&wire.ActorCallRequestIn{ActorId: []string{"123"}, ActionName: "Exclusive", Arguments: []Assignable{FromString("Too-late")}}, handleResult)
	container.Dispatch(&wire.ActorCallRequestIn{ActorId: []string{"234"}, ActionName: "Exclusive", Arguments: []Assignable{FromString("Too-late")}}, handleResult)

	time.Sleep(time.Second)

	checks.Equal(t, []string{
		"123:hello",
		"123:world",
		"234:moon",
		"ERR:CONTAINER_DEACTIVATING",
		"ERR:CONTAINER_DEACTIVATING",
		"CONTAINER-STOPPED",
	}, results, "Results should be as expected")
}
