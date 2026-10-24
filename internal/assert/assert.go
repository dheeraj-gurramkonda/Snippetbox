package assert

import (
	"reflect"
	"testing"
)

func Equal[T any](t *testing.T, got , want T) {
	t.Helper()
    if !isEqual(got, want) {
        t.Errorf("got: %v; want: %v", got, want)
    }
}

func isEqual[T any](got, want T) bool {
	if isNil(got) && isNil(want) {
        return true
    }

    return reflect.DeepEqual(got, want)
}

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

func NotEqual[T comparable](t *testing.T, got, want T) {
    t.Helper()

    if got == want {
        t.Errorf("got: %v; expected values to be different", got)
    }
}

func True(t *testing.T, got bool) {
    t.Helper()

    if !got {
        t.Errorf("got: false; want: true")
    }
}

func False(t *testing.T, got bool) {
    t.Helper()

    if got {
        t.Errorf("got: true; want: false")
    }
}

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