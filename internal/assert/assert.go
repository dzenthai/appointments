package assert

import "testing"

func Equal[T comparable](t *testing.T, actual, expected T) {
	t.Helper()
	if expected != actual {
		t.Errorf("actual: %v, expected: %v\n", actual, expected)
	}
}
