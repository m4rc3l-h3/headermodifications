package assert

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func Contains(t *testing.T, s, contains string) bool {
	t.Helper()

	if !strings.Contains(s, contains) {
		t.Logf("Expected '%s' to contain '%s'", s, contains)
		t.Fail()
		return false
	}
	return true
}

func Containsf(t *testing.T, s, contains, format string, args ...interface{}) bool {
	t.Helper()

	if !strings.Contains(s, contains) {
		t.Logf("Expected '%s' to contain '%s': %s", s, contains, fmt.Sprintf(format, args...))
		t.Fail()
		return false
	}
	return true
}

func NotContains(t *testing.T, s, notContains string) bool {
	t.Helper()

	if strings.Contains(s, notContains) {
		t.Logf("Expected '%s' to NOT contain '%s'", s, notContains)
		t.Fail()
		return false
	}
	return true
}

func Error(t *testing.T, err error) bool {
	t.Helper()

	if err == nil {
		t.Logf("Error is nil, but should have an error")
		t.Fail()

		return false
	}

	return true
}

func NoError(t *testing.T, err error) bool {
	t.Helper()

	if err != nil {
		t.Logf("Error is not nil, but should not have an error")
		t.Fail()

		return false
	}

	return true
}

func Equal(t *testing.T, expect, actual interface{}) bool {
	t.Helper()

	if !reflect.DeepEqual(expect, actual) {
		t.Logf("Expect %v, but got %v", expect, actual)
		t.Fail()

		return false
	}

	return true
}

func Equalf(t *testing.T, expect, actual interface{}, format string, args ...interface{}) bool {
	t.Helper()

	if !reflect.DeepEqual(expect, actual) {
		t.Logf("Expect %v, but got %v: %s", expect, actual, fmt.Sprintf(format, args...))
		t.Fail()

		return false
	}

	return true
}
