package interpreter

import (
	"errors"
	"maps"
)

var (
	errVarAlreadyExists = errors.New("variable already exists")
	errVarNotExists     = errors.New("variable not exists")
	errVarIsImmutable   = errors.New("variable is immutable")
)

type wrappedValue struct {
	Value   Value
	Mutable bool
}

type Env struct {
	store map[string]*wrappedValue
	outer *Env
}

func NewEnv(outer *Env) *Env {
	return &Env{
		store: make(map[string]*wrappedValue),
		outer: outer,
	}
}

func (e *Env) Declare(name string, value Value, mutable bool) error {
	if _, exists := e.store[name]; exists {
		return errVarAlreadyExists
	}
	e.store[name] = &wrappedValue{Value: value, Mutable: mutable}
	return nil
}

func (e *Env) Get(name string) (Value, error) {
	v, exists := e.store[name]
	if exists {
		return v.Value, nil
	}
	if e.outer != nil {
		return e.outer.Get(name)
	}
	return nil, errVarNotExists
}

func (e *Env) Set(name string, value Value) error {
	if v, exists := e.store[name]; exists {
		if !v.Mutable {
			return errVarIsImmutable
		}
		v.Value = value
		return nil
	}
	if e.outer != nil {
		return e.outer.Set(name, value)
	}
	return errVarNotExists
}

func (e *Env) Clone() *Env {
	var outer *Env
	if e.outer != nil {
		outer = e.outer.Clone()
	}
	return &Env{
		store: maps.Clone(e.store),
		outer: outer,
	}
}
