package interpreter

import (
	"fmt"
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

func builtin_class_of(e *Evaluator, this Value, args ...Value) (Value, error) {
	if err := checkArgsLength(1, args); err != nil {
		return nil, err
	}
	if instance, ok := args[0].(*Instance); ok {
		return instance.Class, nil
	}
	return &Null{}, nil
}
