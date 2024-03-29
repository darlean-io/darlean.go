package variant

import (
	"encoding/json"
	"testing"
)

type I interface{ GetName() string }
type T struct{ Name string }

func (t T) GetName() string {
	return t.Name
}

func TestAssignNumber(t *testing.T) {
	int0 := 42
	bytes, _ := json.Marshal(int0)
	var int1 any
	json.Unmarshal(bytes, &int1)
	Check(t, 42.0, int1, "Int unmarshall to any should be ok (albeit converted to float)")
	var int2 int
	Assign(int1, &int2)
	Check(t, 42, int2, "Int assignment via any should be ok (albeit converted to float)")

	var flt32 float32
	Assign(int1, &flt32)
	Check(t, float32(42.0), flt32, "Int assignment via any to float32 should be ok")

	var flt64 float64
	Assign(int1, &flt64)
	Check(t, 42.0, flt64, "Int assignment via any to float64 should be ok")

	Assign(99, &int2)
	Check(t, 99, int2, "Direct int assignment should be ok")
}

func TestAssignString(t *testing.T) {
	str0 := "Hello"
	bytes, _ := json.Marshal(str0)
	var str1 any
	json.Unmarshal(bytes, &str1)
	var str2 string
	Assign(str1, &str2)
	Check(t, "Hello", str2, "String assignment via any should be ok")

	Assign("World", &str2)
	Check(t, "World", str2, "Direct string assignment should be ok")
}

func TestAssignStruct(t *testing.T) {
	val0 := T{Name: "Foo"}
	var val1 T
	Assign(val0, &val1)
	Check(t, val1.GetName(), "Foo", "Direct assignment must be ok")

	var bytes, _ = json.Marshal(val0)
	var val2 any
	json.Unmarshal(bytes, &val2)
	var val3 T
	Assign(val2, &val3)
	Check(t, val3.GetName(), "Foo", "Unmarshalled assignment must be ok")
}

func TestAssignInterface(t *testing.T) {
	val0 := I(T{Name: "Foo"})
	var val1 I
	Assign(val0, &val1)
	Check(t, val1.GetName(), "Foo", "Direct assignment must be ok")

	var bytes, _ = json.Marshal(val0)
	var val2 any
	json.Unmarshal(bytes, &val2)

	var val3 I
	err := Assign(val2, &val3)
	CheckNotNil(t, err, "Unmarshalled assigned to interface is not possible (only to a struct)")

	var val4 T
	Assign(val2, &val4)
	Check(t, val4.GetName(), "Foo", "Unmarshalled assignment to struct must be ok")
}

func TestAssignArray(t *testing.T) {
	val0 := []T{{Name: "Foo"}, {Name: "Bar"}}
	var val1 []T
	Assign(val0, &val1)
	Check(t, val1[0].GetName(), "Foo", "Direct assignment must be ok")
	Check(t, val1[1].GetName(), "Bar", "Direct assignment must be ok")

	var bytes, _ = json.Marshal(val0)
	var val2 any
	json.Unmarshal(bytes, &val2)
	var val3 []T
	Assign(val2, &val3)
	Check(t, val3[0].GetName(), "Foo", "Unmarshalled assignment must be ok")
	Check(t, val3[1].GetName(), "Bar", "Unmarshalled assignment must be ok")
}

func Check(tester *testing.T, expected any, actual any, msg string) {
	ok := expected == actual
	if ok {
		tester.Logf("passed  %s (expected = actual = `%+v`)", msg, actual)
	} else {
		tester.Fatalf("FAILED  %s (expected: `%+v`, actual: `%+v`)", msg, expected, actual)
	}
}

func CheckNotNil(tester *testing.T, actual any, msg string) {
	if actual != nil {
		tester.Logf("passed  %s (actual = `%+v` is not nil)", msg, actual)
	} else {
		tester.Fatalf("FAILED  %s (actual = `%+v` must not be nil)", msg, actual)
	}
}
