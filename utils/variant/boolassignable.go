package variant

import (
	"errors"
	"reflect"
)

type boolVariant bool

func (data boolVariant) AssignTo(target any) error {
	targetVal := reflect.Indirect(reflect.ValueOf(target))
	return data.AssignToReflectValue(&targetVal)
}

func (data boolVariant) AssignToReflectValue(targetVal *reflect.Value) error {
	targetKind := getKind(*targetVal)

	if targetKind == reflect.Bool {
		targetVal.SetBool(bool(data))
		return nil
	}
	return errors.New("variant: target is not a bool")
}

func (data boolVariant) AssignToString() (value string, err error) {
	err = errors.New("variant: cannot assign bool to a string")
	return
}

func (data boolVariant) AssignToBool() (value bool, err error) {
	value = bool(data)
	return
}

func (data boolVariant) AssignToFloat() (value float64, err error) {
	err = errors.New("variant: cannot assign bool to a float")
	return
}

func (data boolVariant) AssignToInt() (value int, err error) {
	err = errors.New("variant: cannot assign bool to an int")
	return
}

func (data boolVariant) AssignToBytes() (value []byte, err error) {
	err = errors.New("variant: cannot assign bool to a byte array")
	return
}

/*
FromBool returns a new [Assignable] that can be assigned to a bool variable.
*/
func FromBool(value bool) Assignable {
	return boolVariant(value)
}
