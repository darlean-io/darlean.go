package core

import "github.com/darlean-io/darlean.go/core/wire"

type Transport interface {
	GetInputChannel() chan *wire.TagsIn
	Send(tags wire.TagsOut) error
}
