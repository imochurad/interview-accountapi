package interview_accountapi

import (
	"fmt"
	"testing"
)

func assertPrimitiveSlices[T string | int | byte](a, b []T) bool {
	if a != nil && b == nil {
		return false
	}

	if a == nil && b != nil {
		return false
	}

	if a == nil && b == nil {
		return true
	}

	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func assertPrimitivePointers[P string | bool | int | int64 | byte](t *testing.T, actual *P, expected *P, varName string) {
	if actual == nil && expected != nil {
		t.Errorf("Actual %s should not be nil", varName)
	}

	if actual != nil && expected == nil {
		t.Errorf("Actual %s payload should be nil", varName)
	}

	if actual != nil && expected != nil {
		if *actual != *expected {
			t.Errorf("%s doesn't match, expected=%s, got=%s", varName,
				fmt.Sprintf("%v", *expected), fmt.Sprintf("%v", *actual))
		}
	}
}
