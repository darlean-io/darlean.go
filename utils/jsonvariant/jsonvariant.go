package jsonvariant

import (
	"github.com/darlean-io/darlean.go/utils/jsonbinary"
	"github.com/darlean-io/darlean.go/utils/variant"
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
