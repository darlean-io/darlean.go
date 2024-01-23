package variant

import (
	"errors"
	"reflect"
)

type floatVariant float64

func (data floatVariant) AssignTo(target any) error {
	targetVal := reflect.Indirect(reflect.ValueOf(target))
	return data.AssignToReflectValue(&targetVal)
}

func (data floatVariant) AssignToReflectValue(targetVal *reflect.Value) error {
	targetKind := getKind(*targetVal)

	if targetKind == reflect.Float64 {
		targetVal.SetFloat(float64(data))
		return nil
	}

	if targetVal.Kind() == reflect.Interface && targetVal.IsZero() {
		targetVal.Set(reflect.ValueOf(data))
		return nil
	}

	return errors.New("variant: target is not a float64")
}

func (data floatVariant) AssignToString() (value string, err error) {
	err = errors.New("variant: cannot assign float to a string")
	return
}

func (data floatVariant) AssignToBool() (value bool, err error) {
	err = errors.New("variant: cannot assign float to a bool")
	return
}

func (data floatVariant) AssignToFloat() (value float64, err error) {
	value = float64(data)
	return
}

func (data floatVariant) AssignToInt() (value int, err error) {
	err = errors.New("variant: cannot assign float to an int")
	return
}

func (data floatVariant) AssignToBytes() (value []byte, err error) {
	err = errors.New("variant: cannot assign float to a byte array")
	return
}

/*
FromFloatNumber returns a new [Assignable] that can be assigned to a float variable.
*/
func FromFloat(value float64) Assignable {
	return floatVariant(value)
}
