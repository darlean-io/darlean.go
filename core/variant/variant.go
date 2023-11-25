package variant

import (
	"errors"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

// Represents a variant value (that can have any type) that may be
// available directly or may first have to be unmarshalled internally.
type Variant interface {
	// Fills `value` with the value currently hold by the variant.
	// The `value` usually is a pointer to the variable that should
	// finally hold the value (like &myVar). The variable itself should
	// be declared as a plain (non-pointer) type like `MyStruct`, `string`,
	// `SomeInterface`` or `[]byte``.
	Get(value any) error

	// Get the value directly (faster) and returns `true` when that was possible. Returns
	// `false` a retry with `Get` should be performed to get the value.
	GetDirect() (any, bool)
}

type baseVariant struct {
	value any
}

// Inspired by: https://go.dev/play/p/u_lBBPeRfz and https://stackoverflow.com/questions/34817188/replace-value-of-interface
func (variant *baseVariant) Get(value any) error {
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Ptr {
		return errors.New("variant: not a pointer")
	}

	val = val.Elem()

	newVal := reflect.Indirect(reflect.ValueOf(variant.value))

	if !val.Type().AssignableTo(newVal.Type()) {
		// Allow assignment to an "any" ("interface {}")
		if val.Type().Kind() != reflect.Interface {
			return errors.New("variant: mismatched types123")
		}
	}

	val.Set(newVal)
	return nil
}

func (variant *baseVariant) GetDirect() (any, bool) {
	return variant.value, true
}

func New(value any) Variant {
	if value == nil {
		return nil
	}

	variant := baseVariant{
		value: value,
	}
	return &variant
}

func Array(values ...any) []Variant {
	result := make([]Variant, len(values))
	for i, v := range values {
		result[i] = New(v)
	}
	return result
}

func Map(m map[string]any) map[string]Variant {
	result := make(map[string]Variant, len(m))
	for k, v := range m {
		result[k] = New(v)
	}
	return result
}

func Assign(source any, target any) error {
	return mapstructure.Decode(source, target)
}
