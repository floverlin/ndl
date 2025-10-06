package evaluator

import (
	"errors"
	"fmt"
	"needle/internal/needle/parser"
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
	VAL_NULL      ValueType = "null"
	VAL_BOOLEAN   ValueType = "boolean"
	VAL_NUMBER    ValueType = "number"
	VAL_STRING    ValueType = "string"
	VAL_FUNCTION  ValueType = "function"
	VAL_METHOD    ValueType = "method"
	VAL_INSTANCE  ValueType = "instance"
	VAL_EXCEPTION ValueType = "exception"
	VAL_CLASS     ValueType = "class"
	VAL_ARRAY     ValueType = "array"
	VAL_MAP       ValueType = "map"
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

func (s *String) Type() ValueType { return VAL_STRING }
func (s *String) Debug() string {
	return fmt.Sprintf("\"%s\"", s.Value)
}

type Function struct {
	FType      FType
	Parameters []string
	Body       []parser.Statement
	Native     NativeFunction
	Closure    *Env
}

func (f *Function) Type() ValueType {
	return VAL_FUNCTION
}
func (f *Function) Debug() string {
	return fmt.Sprintf("<function %p>", f)
}

type Method struct {
	Function      *Function
	This          *Instance
	IsConstructor bool
}

func (m *Method) Type() ValueType {
	return VAL_METHOD
}
func (m *Method) Debug() string {
	return fmt.Sprintf("<function %p>", m)
}

type Class struct {
	Fields       map[string]Value
	Constructors map[string]*Function
	Public       map[string]*Function
	Private      map[string]*Function
	Getters      map[string]*Function
	Setters      map[string]*Function
}

func (c *Class) Type() ValueType { return VAL_FUNCTION }
func (c *Class) Debug() string {
	return fmt.Sprintf("<class %p>", c)
}

type Instance struct {
	Class  *Class
	Fields map[string]Value
}

func (i *Instance) Type() ValueType { return VAL_FUNCTION }
func (i *Instance) Debug() string {
	return fmt.Sprintf("<instance %p of class %p>", i, i.Class)
}

type Exception struct {
	Message    string
	StackTrace []string
}

func (e *Exception) Type() ValueType { return VAL_EXCEPTION }
func (e *Exception) Debug() string {
	return fmt.Sprintf("<exception %p>", e)
}



type Array struct {
	Elements []Value
}

func (a *Array) Type() ValueType { return VAL_ARRAY }
func (a *Array) Debug() string {
	return fmt.Sprintf("<array %p>", a)
}

type Map struct {
	Pairs *HashTable
}

func (m *Map) Type() ValueType { return VAL_MAP }
func (m *Map) Debug() string {
	return fmt.Sprintf("<map %p>", m)
}

type HashTable struct {
	boolMap map[bool]Value
	numMap  map[float64]Value
	strMap  map[string]Value
}

func NewHashTable() *HashTable {
	return &HashTable{
		boolMap: map[bool]Value{},
		numMap:  map[float64]Value{},
		strMap:  map[string]Value{},
	}
}

func (ht *HashTable) Get(key Value) (Value, error) {
	switch key := key.(type) {
	case *Boolean:
		if v, ok := ht.boolMap[key.Value]; ok {
			return v, nil
		}
		return nil, errors.New("missing key")
	case *Number:
		if v, ok := ht.numMap[key.Value]; ok {
			return v, nil
		}
		return nil, errors.New("missing key")
	case *String:
		if v, ok := ht.strMap[key.Value]; ok {
			return v, nil
		}
		return nil, errors.New("missing key")
	default:
		return nil, errors.New("unhashable type")
	}
}

func (ht *HashTable) Set(key Value, value Value) (bool, error) {
	switch key := key.(type) {
	case *Boolean:
		_, ok := ht.boolMap[key.Value]
		ht.boolMap[key.Value] = value
		return ok, nil
	case *Number:
		_, ok := ht.numMap[key.Value]
		ht.numMap[key.Value] = value
		return ok, nil
	case *String:
		_, ok := ht.strMap[key.Value]
		ht.strMap[key.Value] = value
		return ok, nil
	default:
		return false, errors.New("unhashable type")
	}
}
