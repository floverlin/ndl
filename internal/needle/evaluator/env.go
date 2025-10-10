package evaluator

import (
	"errors"
	"maps"
)

var (
	errVarAlreadyExists = errors.New("variable already exists")
	errVarNotExists     = errors.New("variable not exists")
)

type Globals struct {
	Null  *Null
	True  *Boolean
	False *Boolean
}

func newGlobals() *Globals {
	return &Globals{
		Null:  &Null{},
		True:  &Boolean{Value: true},
		False: &Boolean{Value: false},
	}
}

type Env struct {
	store   map[string]Value
	outer   *Env
	this    Value
	globals *Globals
}

// outer can be nil, but root env must have global outer
func NewEnv(outer *Env) *Env {
	var g *Globals
	if outer != nil {
		g = outer.globals
	} else {
		g = newGlobals()
	}
	return &Env{
		store:   make(map[string]Value),
		outer:   outer,
		globals: g,
	}
}

func (e *Env) Declare(name string, value Value) error {
	if _, exists := e.store[name]; exists {
		return errVarAlreadyExists
	}
	e.store[name] = value
	return nil
}

func (e *Env) Get(name string) (Value, error) {
	v, exists := e.store[name]
	if exists {
		return v, nil
	}
	if e.outer != nil {
		return e.outer.Get(name)
	}
	return nil, errVarNotExists
}

func (e *Env) Set(name string, value Value) error {
	if _, exists := e.store[name]; exists {
		e.store[name] = value
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
		store:   maps.Clone(e.store),
		outer:   outer,
		globals: e.globals,
	}
}

func (e *Env) GetThis() Value {
	if e.this != nil {
		return e.this
	}
	if e.outer != nil {
		return e.outer.GetThis()
	}
	return nil
}

func (e *Env) SetThis(this Value) {
	e.this = this
}
