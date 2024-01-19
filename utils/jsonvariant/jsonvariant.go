package jsonvariant

import (
	"reflect"

	"github.com/darlean-io/darlean.go/utils/jsonbinary"
	"github.com/darlean-io/darlean.go/utils/variant"
)

type jsonVariant []byte

func FromJson(data []byte) variant.Assignable {
	return jsonVariant(data)
}

func (data jsonVariant) AssignTo(target any) error {
	return jsonbinary.Deserialize(data, target)
}

func (data jsonVariant) AssignToReflectValue(targetVal *reflect.Value) error {
	return jsonbinary.Deserialize(data, targetVal)
}

func (data jsonVariant) MarshalJSON() ([]byte, error) {
	return data, nil
}

func (data jsonVariant) AssignToBool() (value bool, err error) {
	err = data.AssignTo(&value)
	return
}

func (data jsonVariant) AssignToFloat() (value float64, err error) {
	err = data.AssignTo(&value)
	return
}

func (data jsonVariant) AssignToInt() (value int, err error) {
	err = data.AssignTo(&value)
	return
}

func (data jsonVariant) AssignToBytes() (value []byte, err error) {
	err = data.AssignTo(&value)
	return
}

func (data jsonVariant) AssignToString() (value string, err error) {
	err = data.AssignTo(&value)
	return
}
