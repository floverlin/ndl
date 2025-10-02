package interpreter

import (
	"errors"
	"fmt"
	"maps"
	"time"
)

type NativeFunction func(e *Evaluator, this Value, args ...Value) (Value, error)

func LoadBuiltins(env *Env) {
	for name, builtin := range builtins {
		fun := &Function{
			FType:  F_NATIVE,
			Native: builtin,
		}
		env.Declare(name, fun, false)
	}
	env.Declare("Exception", Exception, false)
}

// func NewException(message string) Value {
// 	excp := &Instance{
// 		Class:  nil,
// 		Fields: nil,
// 	}
// }

var Exception = &Class{
	Fields: map[string]Value{
		"message": &Null{},
	},
	Constructors: map[string]*Function{
		"new": {
			FType: F_NATIVE,
			Native: func(
				e *Evaluator,
				this Value,
				args ...Value,
			) (Value, error) {
				if err := checkArgsLength(1, args); err != nil {
					return nil, err
				}
				message, ok := args[0].(*String)
				if !ok {
					return nil, errors.New("expected string")
				}
				this.(*Instance).Fields["message"] = message
				return nil, nil
			},
		},
	},
}

func LoadKlass(env *Env) {
	clock := methods["hello"]
	new := ctors["new"]
	Klass := &Class{
		Methods: map[string]*Function{
			"clock": {
				FType:  F_NATIVE,
				Native: clock,
			},
		},
		Fields: map[string]Value{
			"name": &String{Value: "Lin"},
		},
		Constructors: map[string]*Function{
			"new": {
				FType:  F_NATIVE,
				Native: new,
			},
		},
	}
	env.Declare("Klass", Klass, false)
	klass := &Instance{
		Class:  Klass,
		Fields: maps.Clone(Klass.Fields),
	}
	env.Declare("klass", klass, false)
}

var builtins = map[string]NativeFunction{
	"clock": builtin_clock,
}
var methods = map[string]NativeFunction{
	"hello": builtin_hello,
}
var ctors = map[string]NativeFunction{
	"new": builtin_ctor,
}

func checkArgsLength(length int, args []Value) error {
	if length < 0 {
		return nil
	}
	if len(args) != length {
		return fmt.Errorf(
			"expected %d arguments, got %d",
			length,
			len(args),
		)
	}
	return nil
}

func builtin_clock(e *Evaluator, this Value, args ...Value) (Value, error) {
	if err := checkArgsLength(0, args); err != nil {
		return nil, err
	}
	t := float64(time.Now().UnixNano()) / float64(time.Second)
	return &Number{Value: t}, nil
}

func builtin_hello(e *Evaluator, this Value, args ...Value) (Value, error) {
	if err := checkArgsLength(0, args); err != nil {
		return nil, err
	}
	name := this.(*Instance).Fields["name"].(*String)
	fmt.Printf("Hello, %s!\n", name.Value)
	return &Null{}, nil
}
func builtin_ctor(e *Evaluator, this Value, args ...Value) (Value, error) {
	if err := checkArgsLength(1, args); err != nil {
		return nil, err
	}
	this.(*Instance).Fields["name"] = args[0].(*String)
	return nil, nil
}
