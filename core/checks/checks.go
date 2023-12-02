package checks

import (
	"bytes"
	"encoding/json"
	"testing"
)

func Equal(tester *testing.T, expected any, actual any, msg string) {
	exp, err := json.Marshal(expected)
	if err != nil {
		panic(err)
	}
	act, err := json.Marshal(actual)
	if err != nil {
		panic(err)
	}
	ok := bytes.Compare(exp, act) == 0
	if ok {
		tester.Logf("passed  %s (expected = actual = `%+v`)", msg, actual)
	} else {
		tester.Fatalf("FAILED  %s (expected: `%+v`, actual: `%+v`)", msg, expected, actual)
	}
}

func IsNotNil(tester *testing.T, actual any, msg string) {
	if actual != nil {
		tester.Logf("passed  %s (actual = `%+v` is not nil)", msg, actual)
	} else {
		tester.Fatalf("FAILED  %s (actual = `%+v` must not be nil)", msg, actual)
	}
}
