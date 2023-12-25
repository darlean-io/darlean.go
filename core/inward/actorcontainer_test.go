package inward

import (
	"fmt"
	"testing"
	"time"

	"github.com/darlean-io/darlean.go/core/wire"

	"github.com/darlean-io/darlean.go/utils/checks"
)

func TestActorContainer(t *testing.T) {
	wrapperFactory := func(id []string) InstanceWrapper {
		wrapper := TestActorWrapper{
			id: id[0],
		}
		return &wrapper
	}

	var results []string

	container := NewStandardActorContainer(false, GetTestActionDefs(), wrapperFactory, func() {
		results = append(results, "CONTAINER-STOPPED")
	})

	handleResult := func(result any, err error) {
		if err != nil {
			results = append(results, fmt.Sprintf("ERR:%v", err))
		} else {
			results = append(results, fmt.Sprintf("%v", result))
		}
	}

	container.Dispatch(&wire.ActorCallRequest{ActorId: []string{"123"}, ActionName: "Exclusive", Arguments: []any{"Hello"}}, handleResult)
	container.Dispatch(&wire.ActorCallRequest{ActorId: []string{"123"}, ActionName: "Exclusive", Arguments: []any{"World"}}, handleResult)
	container.Dispatch(&wire.ActorCallRequest{ActorId: []string{"234"}, ActionName: "Exclusive", Arguments: []any{"Moon"}}, handleResult)

	time.Sleep(SLEEP_BASIS * 5)

	container.Stop()

	time.Sleep(SLEEP_BASIS_HALF)

	container.Dispatch(&wire.ActorCallRequest{ActorId: []string{"123"}, ActionName: "Exclusive", Arguments: []any{"Too-late"}}, handleResult)
	container.Dispatch(&wire.ActorCallRequest{ActorId: []string{"234"}, ActionName: "Exclusive", Arguments: []any{"Too-late"}}, handleResult)

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
