package pkg_test

import (
	"errors"
	"needle/internal/pkg"
	"testing"
)

func TestStack(t *testing.T) {
	s := pkg.NewStack[*int]()
	var v *int
	var err error

	if s == nil {
		t.Errorf("unexpected nil")
	}

	v, err = s.Pop()
	if v != nil {
		t.Errorf("wrong zero value")
	}
	if err == nil {
		t.Errorf("unexpected nil error")
	}
	if !errors.Is(err, pkg.ErrEmpty) {
		t.Errorf("wrong error")
	}

	var a1, a2 int = 1, 2
	s.Push(&a1)
	s.Push(&a2)

	if s.Length() != 2 {
		t.Errorf("wrong stack length")
	}

	v, err = s.Pop()
	if *v != 2 {
		t.Errorf("wrong pop value")
	}
	if err != nil {
		t.Errorf("unexpected error")
	}
	if s.Length() != 1 {
		t.Errorf("wrong stack length")
	}
}
