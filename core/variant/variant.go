package variant

import (
	"github.com/mitchellh/mapstructure"
)

func Assign(source any, target any) error {
	return mapstructure.Decode(source, target)
}
