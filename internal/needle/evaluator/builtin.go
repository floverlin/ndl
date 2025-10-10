package evaluator

import (
	"fmt"
	"math/rand/v2"
	"time"
)

type NativeFunction func(e *Evaluator, this Value, args ...Value) Value

func LoadBuiltins(e *Evaluator) {
	for name, builtin := range newBuiltins() {
		fun := &Function{
			FType:  F_NATIVE,
			Native: builtin,
		}
		e.env.Declare(name, fun)
	}
}

func newBuiltins() map[string]NativeFunction {
	builtins := map[string]NativeFunction{
		"clock":    coverNative(builtin_clock, 0),
		"class_of": coverNative(builtin_class_of, 1),
		"random":   coverNative(builtin_random, 0),
	}
	return builtins
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

func coverNative(f NativeFunction, a int) NativeFunction {
	return func(
		e *Evaluator,
		this Value,
		args ...Value,
	) Value {
		err := CheckArgsLength(a, args)
		if err != nil {
			e.ThrowException("%s", err.Error())
		}
		return f(e, this, args...)
	}
}

func builtin_clock(e *Evaluator, this Value, args ...Value) Value {
	t := float64(time.Now().UnixNano()) / float64(time.Second)
	return &Number{Value: t}
}

func builtin_class_of(e *Evaluator, this Value, args ...Value) Value {
	if instance, ok := args[0].(*Instance); ok {
		return instance.Class
	}
	return &Null{}
}

func builtin_random(e *Evaluator, this Value, args ...Value) Value {
	r := rand.Float64()
	return &Number{Value: r}
}
