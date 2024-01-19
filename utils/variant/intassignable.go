package variant

import (
	"errors"
	"reflect"
)

type intVariant int

func (data intVariant) AssignTo(target any) error {
	targetVal := reflect.Indirect(reflect.ValueOf(target))
	return data.AssignToReflectValue(&targetVal)
}

func (data intVariant) AssignToReflectValue(targetVal *reflect.Value) error {
	targetKind := getKind(*targetVal)

	if targetKind == reflect.Int {
		targetVal.SetInt(int64(data))
		return nil
	}
	return errors.New("variant: target is not an int")
}

func (data intVariant) AssignToString() (value string, err error) {
	err = errors.New("variant: cannot assign int to a string")
	return
}

func (data intVariant) AssignToBool() (value bool, err error) {
	err = errors.New("variant: cannot assign int to a bool")
	return
}

func (data intVariant) AssignToInt() (value int, err error) {
	value = int(data)
	return
}

func (data intVariant) AssignToFloat() (value float64, err error) {
	err = errors.New("variant: cannot assign int to a float")
	return
}

func (data intVariant) AssignToBytes() (value []byte, err error) {
	err = errors.New("variant: cannot assign int to a byte array")
	return
}

/*
FromIntNumber returns a new [Assignable] that can be assigned to an integer variable.
*/
func Fromint64(value int64) Assignable {
	return intVariant(value)
}
