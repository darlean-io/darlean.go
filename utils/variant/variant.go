package variant

import "reflect"

/*
Assignable represents an internal value (or representation thereof) that can be assigned
to some variable.
Depending on the type of the variable, conversions can be done, like
transforming a map of keys and values to the proper struct format or
parsing an internal buffer of json data.
*/
type Assignable interface {
	// AssignTo assigns the internal value to a target variable. It is important to
	// pass a pointer to the target variable, not the variable itself.
	// Returns an error when the internal value cannot be assigned to the target
	// because of a type mismatch.
	AssignTo(target any) error
	// AssignToReflectValue assigns the internal value to a reflect value.
	// Returns an error when the internal value cannot be assigned to the target
	// because of a type mismatch.
	AssignToReflectValue(targetVal *reflect.Value) error
	// AssignToString returns the internal value when it is a string.
	// Returns an error when the internal value cannot be assigned to a string
	// because of a type mismatch.
	AssignToString() (string, error)
	// AssignToBool returns the internal value when it is a bool.
	// Returns an error when the internal value cannot be assigned to a bool
	// because of a type mismatch.
	AssignToBool() (bool, error)
	// AssignToBytes returns the internal value when it is a []byte.
	// Returns an error when the internal value cannot be assigned to a byte[]
	// because of a type mismatch.
	AssignToBytes() ([]byte, error)
	// AssignToFloat returns the internal value when it is a float.
	// Returns an error when the internal value cannot be assigned to a float
	// because of a type mismatch.
	AssignToFloat() (float64, error)
	// AssignToInt returns the internal value when it is an int.
	// Returns an error when the internal value cannot be assigned to the target
	// because of a type mismatch.
	AssignToInt() (int, error)
}

/*
Assign assigns a source value to a target variable (which must be a pointer to the actual value).
When the source satisfies [Assignable], the [Assignable.AssignTo] is used. Otherwise, the functionality
from the mapstruct library is used.
*/
func Assign(source any, target any) error {
	assignable, supported := source.(Assignable)
	if supported {
		return assignable.AssignTo(target)
	}
	return Decode(source, target)
}
