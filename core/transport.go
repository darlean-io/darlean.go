package core

import "core/wire"

type Transport interface {
	GetInputChannel() chan *wire.Tags
	Send(tags wire.Tags) error
}
