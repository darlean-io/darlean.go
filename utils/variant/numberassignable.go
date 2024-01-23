package variant

import (
	"errors"
	"reflect"
)

type numberVariant float64

func (data numberVariant) AssignTo(target any) error {
	targetVal := reflect.Indirect(reflect.ValueOf(target))
	return data.AssignToReflectValue(&targetVal)

}

func (data numberVariant) AssignToReflectValue(targetVal *reflect.Value) error {
	targetKind := getKind(*targetVal)

	if targetKind == reflect.Float64 {
		targetVal.SetFloat(float64(data))
		return nil
	}
	if targetKind == reflect.Int {
		targetVal.SetInt(int64(float64(data)))
		return nil
	}

	if targetVal.Kind() == reflect.Interface && targetVal.IsZero() {
		targetVal.Set(reflect.ValueOf(data))
		return nil
	}

	return errors.New("variant: target is not a number")
}

func (data numberVariant) AssignToString() (value string, err error) {
	err = errors.New("variant: cannot assign number to a string")
	return
}

func (data numberVariant) AssignToBool() (value bool, err error) {
	err = errors.New("variant: cannot assign number to a bool")
	return
}

func (data numberVariant) AssignToFloat() (value float64, err error) {
	value = float64(data)
	return
}

func (data numberVariant) AssignToInt() (value int, err error) {
	if float64(data) != float64(int(float64(data))) {
		err = errors.New("variant: cannot assign float number to an int")
	}
	value = int(data)
	return
}

func (data numberVariant) AssignToBytes() (value []byte, err error) {
	err = errors.New("variant: cannot assign number to a byte array")
	return
}

/*
FromFloatNumber returns a new [Assignable] that can be assigned to a float variable, or
to an int variable when value does not have a fractional value.
*/
func FromFloatNumber(value float64) Assignable {
	return numberVariant(value)
}

/*
FromIntNumber returns a new [Assignable] that can be assigned to an integer or float variable.
*/
func FromIntNumber(value int) Assignable {
	return numberVariant(value)
}
