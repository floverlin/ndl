package evaluator

import (
	"fmt"
	"math/rand/v2"
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
}

var builtins = map[string]NativeFunction{
	"clock":    builtin_clock,
	"class_of": builtin_class_of,
	"random":   builtin_random,
}

func CheckArgsLength(length int, args []Value) error {
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
	if err := CheckArgsLength(0, args); err != nil {
		return nil, err
	}
	t := float64(time.Now().UnixNano()) / float64(time.Second)
	return &Number{Value: t}, nil
}

func builtin_class_of(e *Evaluator, this Value, args ...Value) (Value, error) {
	if err := CheckArgsLength(1, args); err != nil {
		return nil, err
	}
	if instance, ok := args[0].(*Instance); ok {
		return instance.Class, nil
	}
	return &Null{}, nil
}

func builtin_random(e *Evaluator, this Value, args ...Value) (Value, error) {
	if err := CheckArgsLength(0, args); err != nil {
		return nil, err
	}
	r := rand.Float64()
	return &Number{Value: r}, nil
}
