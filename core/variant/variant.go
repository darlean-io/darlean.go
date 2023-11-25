package variant

import (
	"core/jsonbinary"

	"github.com/mitchellh/mapstructure"
)

type Assignable interface {
	AssignTo(target any) error
}

func Assign(source any, target any) error {
	assignable, supported := source.(Assignable)
	if supported {
		return assignable.AssignTo(target)
	}
	return mapstructure.Decode(source, target)
}

type jsonData []byte

func NewJsonAssignable(data []byte) Assignable {
	return jsonData(data)
}

func (data jsonData) AssignTo(target any) error {
	return jsonbinary.Deserialize(data, target)
}

func (data jsonData) MarshalJSON() ([]byte, error) {
	return data, nil
}
