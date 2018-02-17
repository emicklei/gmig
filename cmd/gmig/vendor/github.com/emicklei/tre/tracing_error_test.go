package tre

import (
	"errors"
	"testing"
)

func TestTracingError(t *testing.T) {
	if len(rootPath) == 0 {
		t.Fail()
	}
	e := New(propError(), "prop failed", "ik", "Koen").(*TracingError)
	if got, want := len(e.callTrace), 2; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := Cause(e).Error(), "fail 1"; got != want {
		t.Errorf("got %v want %v", got, want)
	}

}

func propError() error {
	return New(giveError(), "give failed", "a", 42)
}

func giveError() error {
	return errors.New("fail 1")
}

func TestEmptyTracingError(t *testing.T) {
	e := New(errors.New("empty"), "empty").(*TracingError)
	ctx := e.LoggingContext()
	if ctx["err"] != e.cause {
		t.Error("err expected")
	}
	if ctx["err"] != e.cause {
		t.Error("err expected")
	}
	if ctx["msg"] != "empty" {
		t.Error("empty expected")
	}
}

func TestLengthOfLargestMatchingPrefix(t *testing.T) {
	for _, each := range []struct {
		s1 string
		s2 string
		i  int
	}{
		{"a", "a", 1},
		{"", "a", 0},
		{"", "", 0},
		{"a", "", 0},
		{"abc", "abc", 3},
		{"abc", "ab c", 2},
	} {
		if got, want := lengthOfLargestMatchingPrefix(each.s1, each.s2), each.i; got != want {
			t.Errorf("got %v want %v", got, want)
		}
	}
}
