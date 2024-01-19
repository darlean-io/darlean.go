package variant

import (
	"testing"
)

func TestAssignStringVariant(t *testing.T) {
	v := FromString("Hello")
	v1, _ := v.AssignToString()
	Check(t, "Hello", v1, "Assign to specific type")

	var v2 string
	v.AssignTo(&v2)
	Check(t, "Hello", v2, "Assign to any type")

	_, err3 := v.AssignToBool()
	CheckNotNil(t, err3, "Assign to wrong specific type should return error")

	var v4 bool
	err4 := v.AssignTo(&v4)
	CheckNotNil(t, err4, "Assign to wrong any type")
}
