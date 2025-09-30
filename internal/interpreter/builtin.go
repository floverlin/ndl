package interpreter

type NativeFunction func(e *Evaluator, args ...Value) (Value, error)
