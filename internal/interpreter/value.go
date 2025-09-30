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

const (
	F_SCRIPT   FType = "script"
	F_FUNCTION FType = "function"
	F_NATIVE   FType = "native"
	F_METHOD   FType = "method"
)

const (
	VAL_NULL     ValueType = "null"
	VAL_BOOLEAN  ValueType = "boolean"
	VAL_NUMBER   ValueType = "number"
	VAL_STRING   ValueType = "string"
	VAL_FUNCTION ValueType = "function"
)

type Null struct{}

func (n *Null) Type() ValueType {
	return VAL_NULL
}
func (n *Null) Debug() string {
	return "null"
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ValueType {
	return VAL_BOOLEAN
}
func (b *Boolean) Debug() string {
	return strconv.FormatBool(b.Value)
}

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

func (s *String) Type() ValueType {
	return VAL_STRING
}
func (s *String) Debug() string {
	return s.Value
}

type Function struct {
	FType      FType
	Closure    *Env
	Body       []parser.Statement
	Parameters []string
}

func (f *Function) Type() ValueType {
	return VAL_FUNCTION
}
func (f *Function) Debug() string {
	return "<function>"
}
