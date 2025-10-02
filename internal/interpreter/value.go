package interpreter

import (
	"needle/internal/parser"
	"strconv"
)

type ValueType string

type FType string

type Value interface {
	Type() ValueType
	Debug() string
}

type SignalType string

type Signal interface {
	Signal() SignalType
}

const (
	F_FUNCTION FType = "function"
	F_NATIVE   FType = "native"
)

const (
	SIG_RETURN   SignalType = "return"
	SIG_BREAK    SignalType = "break"
	SIG_CONTINUE SignalType = "continue"
)

const (
	VAL_NULL     ValueType = "null"
	VAL_BOOLEAN  ValueType = "boolean"
	VAL_NUMBER   ValueType = "number"
	VAL_STRING   ValueType = "string"
	VAL_FUNCTION ValueType = "function"
	VAL_METHOD   ValueType = "method"
	VAL_INSTANCE ValueType = "instance"
	VAL_CLASS    ValueType = "class"
)

type ReturnSignal struct {
	Value Value
}

func (rs *ReturnSignal) Signal() SignalType { return SIG_RETURN }

type BreakSignal struct {
	Value Value
}

func (bs *BreakSignal) Signal() SignalType { return SIG_RETURN }

type ContinueSignal struct {
	Value Value
}

func (cs *ContinueSignal) Signal() SignalType { return SIG_RETURN }

type Null struct{}

func (n *Null) Type() ValueType { return VAL_NULL }
func (n *Null) Debug() string   { return "null" }

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ValueType { return VAL_BOOLEAN }
func (b *Boolean) Debug() string   { return strconv.FormatBool(b.Value) }

type Number struct {
	Value float64
}

func (n *Number) Type() ValueType {
	return VAL_NUMBER
}
func (n *Number) Debug() string {
	return strconv.FormatFloat(n.Value, 'g', -1, 64)
}

type String struct {
	Value string
}

func (s *String) Type() ValueType { return VAL_STRING }
func (s *String) Debug() string   { return s.Value }

type Function struct {
	FType      FType
	Parameters []string
	Body       []parser.Statement
	Native     *NativeFunction
	Closure    *Env
}

func (f *Function) Type() ValueType {
	return VAL_FUNCTION
}
func (f *Function) Debug() string { return "<function>" }

type Method struct {
	Function      *Function
	This          *Instance
	IsConstructor bool
}

func (m *Method) Type() ValueType {
	return VAL_METHOD
}
func (m *Method) Debug() string { return "<function>" }

type Class struct {
	Methods      map[string]*Function
	Fields       map[string]Value
	Constructors map[string]*Function
}

func (c *Class) Type() ValueType { return VAL_FUNCTION }
func (c *Class) Debug() string   { return "<class>" }

type Instance struct {
	Class  *Class
	Fields map[string]Value
}

func (i *Instance) Type() ValueType { return VAL_FUNCTION }
func (i *Instance) Debug() string   { return "<instance>" }
