package jsonvariant

import (
	"core/jsonbinary"
	"core/variant"
)

type jsonData []byte

func NewJsonAssignable(data []byte) variant.Assignable {
	return jsonData(data)
}

func (data jsonData) AssignTo(target any) error {
	return jsonbinary.Deserialize(data, target)
}

func (data jsonData) MarshalJSON() ([]byte, error) {
	return data, nil
}
