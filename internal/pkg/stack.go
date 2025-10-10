package pkg

import (
	"errors"
	"slices"
)

var ErrEmpty = errors.New("empty")

type Stack[T any] struct {
	stack []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{
		stack: make([]T, 0),
	}
}

func (s *Stack[T]) Push(value T) {
	s.stack = append(s.stack, value)
}

func (s *Stack[T]) Pop() (T, error) {
	if len(s.stack) == 0 {
		var zero T
		return zero, ErrEmpty
	}
	value := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	if len(s.stack) <= cap(s.stack) {
		n := make([]T, len(s.stack)+int(len(s.stack)/4))
		copy(n, s.stack)
		s.stack = n
	}
	return value, nil
}

func (s *Stack[T]) Length() int {
	return len(s.stack)
}

func (s *Stack[T]) Shot() []T {
	return slices.Clone(s.stack)
}
