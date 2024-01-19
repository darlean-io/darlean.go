package variant

import (
	"fmt"
)

func ExampleFromString() {
	v := FromString("Hello")

	v1, _ := v.AssignToString()
	fmt.Printf("Via AssignToString: %s\n", v1)

	var v2 string
	v.AssignTo(&v2)
	fmt.Printf("Via AssignTo: %s\n", v1)

	_, err3 := v.AssignToBool()
	if err3 != nil {
		fmt.Println("AssignToBool gives error")
	}

	var v4 bool
	err4 := v.AssignTo(&v4)
	if err4 != nil {
		fmt.Println("AssignTo a bool gives error")
	}

	// Output:
	// Via AssignToString: Hello
	// Via AssignTo: Hello
	// AssignToBool gives error
	// AssignTo a bool gives error
}
