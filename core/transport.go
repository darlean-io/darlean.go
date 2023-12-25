package core

import "github.com/darlean-io/darlean.go/core/wire"

type Transport interface {
	GetInputChannel() chan *wire.Tags
	Send(tags wire.Tags) error
}
