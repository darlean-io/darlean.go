package variant

import (
	"errors"
	"reflect"
)

type stringVariant string

func (data stringVariant) AssignTo(target any) error {
	targetVal := reflect.Indirect(reflect.ValueOf(target))
	return data.AssignToReflectValue(&targetVal)
}

func (data stringVariant) AssignToReflectValue(targetVal *reflect.Value) error {
	targetKind := getKind(*targetVal)

	if targetKind == reflect.String {
		targetVal.SetString(string(data))
		return nil
	}

	return errors.New("variant: target is not a string")
}

func (data stringVariant) AssignToString() (value string, err error) {
	value = string(data)
	return
}

func (data stringVariant) AssignToBool() (value bool, err error) {
	err = errors.New("variant: cannot assign string to a bool")
	return
}

func (data stringVariant) AssignToFloat() (value float64, err error) {
	err = errors.New("variant: cannot assign string to a float")
	return
}

func (data stringVariant) AssignToInt() (value int, err error) {
	err = errors.New("variant: cannot assign string to an int")
	return
}

func (data stringVariant) AssignToBytes() (value []byte, err error) {
	err = errors.New("variant: cannot assign string to a byte array")
	return
}

/*
FromString returns a new [Assignable] that can be assigned to a string variable.
*/
func FromString(value string) Assignable {
	return stringVariant(value)
}
