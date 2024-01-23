package variant

import (
	"errors"
	"reflect"
)

type bytesVariant []byte

func (data bytesVariant) AssignTo(target any) error {
	targetVal := reflect.Indirect(reflect.ValueOf(target))
	return data.AssignToReflectValue(&targetVal)
}

func (data bytesVariant) AssignToReflectValue(targetVal *reflect.Value) error {
	targetKind := getKind(*targetVal)

	if targetKind == reflect.Array && targetVal.Type().Elem().Kind() == reflect.Uint8 {
		targetVal.SetBytes([]byte(data))
		return nil
	}

	if targetVal.Kind() == reflect.Interface && targetVal.IsZero() {
		targetVal.Set(reflect.ValueOf(data))
		return nil
	}

	return errors.New("variant: target is not a []byte")
}

func (data bytesVariant) AssignToString() (value string, err error) {
	err = errors.New("variant: cannot assign []byte to a string")
	return
}

func (data bytesVariant) AssignToBool() (value bool, err error) {
	err = errors.New("variant: cannot assign []byte to a bool")
	return
}

func (data bytesVariant) AssignToFloat() (value float64, err error) {
	err = errors.New("variant: cannot assign []byte to a float")
	return
}

func (data bytesVariant) AssignToInt() (value int, err error) {
	err = errors.New("variant: cannot assign []byte] to an int")
	return
}

func (data bytesVariant) AssignToBytes() (value []byte, err error) {
	value = []byte(data)
	return
}

/*
FromBytes returns a new [Assignable] that can be assigned to a []byte variable.
*/
func FromBytes(value []byte) Assignable {
	return bytesVariant(value)
}
