package testings

import (
	"errors"
	"testing"
)

// ErrCheck is a function to check errors in tests
type ErrCheck func(*testing.T, error)

// Ok tests if there is no error
func Ok(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

// ErrorIs checks if an error is of a given type
func ErrorIs(want error) ErrCheck {
	return func(t *testing.T, err error) {
		t.Helper()
		if !errors.Is(err, want) {
			t.Fatalf("err is %v, but want %v", err, want)
		}
	}
}

// ErrorSays checks if an error content is expected
func ErrorSays(str string) ErrCheck {
	return func(t *testing.T, err error) {
		t.Helper()
		if err == nil {
			t.Fatalf("err is nil, but want %v", str)
		}
		if errStr := err.Error(); errStr != str {
			t.Fatalf("err says %v, but want %v", errStr, str)
		}
	}
}
