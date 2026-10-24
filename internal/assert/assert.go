package assert

import (
	"reflect"
	"testing"
)

// Equal compares values deeply so tests can check slices, structs, and other
// non-comparable values in addition to ordinary scalars.
func Equal[T any](t *testing.T, got , want T) {
	t.Helper()
    if !isEqual(got, want) {
        t.Errorf("got: %v; want: %v", got, want)
    }
}

// isEqual treats nil-able values with nil underlying contents as equal before
// delegating structural comparison to reflect.DeepEqual.
func isEqual[T any](got, want T) bool {
	if isNil(got) && isNil(want) {
        return true
    }

    return reflect.DeepEqual(got, want)
}

// isNil safely checks nil-capable kinds without calling IsNil on scalar values.
func isNil(v any) bool {
	if v == nil {
        return true
    }

    rv := reflect.ValueOf(v)
    switch rv.Kind() {
    case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
        return rv.IsNil()
    }

    return false
}

// NotEqual fails the test when got and want are the same comparable value.
func NotEqual[T comparable](t *testing.T, got, want T) {
    t.Helper()

    if got == want {
        t.Errorf("got: %v; expected values to be different", got)
    }
}

// True fails the test unless got is true.
func True(t *testing.T, got bool) {
    t.Helper()

    if !got {
        t.Errorf("got: false; want: true")
    }
}

// False fails the test unless got is false.
func False(t *testing.T, got bool) {
    t.Helper()

    if got {
        t.Errorf("got: true; want: false")
    }
}

// Nil and NotNil check whether an interface contains a nil value.
func Nil(t *testing.T, got any) {
    t.Helper()

    if got != nil {
        t.Errorf("got: %v; want: nil", got)
    }
}

func NotNil(t *testing.T, got any) {
    t.Helper()

    if got == nil {
        t.Errorf("got: nil; want: non-nil")
    }
}