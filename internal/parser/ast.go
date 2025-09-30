package parser

import (
	"fmt"
	"strconv"
	"strings"
)

type Operator = string

const (
	OP_PLUS  Operator = "+"
	OP_MINUS Operator = "-"
	OP_STAR  Operator = "*"
	OP_SLASH Operator = "/"
	OP_EQ    Operator = "=="
	OP_NE    Operator = "!="
	OP_IS    Operator = "==="
	OP_ISNT  Operator = "!=="
	OP_LT    Operator = "<"
	OP_LE    Operator = "<="
	OP_GT    Operator = ">"
	OP_GE    Operator = ">="
	OP_OR    Operator = "or"
	OP_AND   Operator = "and"
	OP_NOT   Operator = "not"
)

type Node interface {
	fmt.Stringer
	Node()
}

type Statement interface {
	Node
	Statement()
}

type Expression interface {
	Node
	Expression()
}

type Script struct {
	Statements []Statement
}

func (s *Script) Node() {}
func (s *Script) String() string {
	var str strings.Builder
	for _, stmt := range s.Statements {
		str.WriteString(stmt.String())
		str.WriteByte('\n')
	}
	return str.String()
}

// == Statements ==

type Block struct {
	Statements []Statement
}

func (b *Block) Node()      {}
func (b *Block) Statement() {}
func (b *Block) String() string {
	var str strings.Builder
	str.WriteString("{ ")
	for _, stmt := range b.Statements {
		str.WriteString(stmt.String())
		str.WriteByte(' ')
	}
	str.WriteByte('}')
	return str.String()
}

type IfStatement struct {
	Condition Expression
	Then      Statement
	Else      Statement
}

func (is *IfStatement) Node()      {}
func (is *IfStatement) Statement() {}
func (is *IfStatement) String() string {
	return fmt.Sprintf(
		"if (%s) %s else %s",
		is.Condition,
		is.Then,
		is.Else,
	)
}

type WhileStatement struct {
	Condition Expression
	Do        Statement
}

func (ws *WhileStatement) Node()      {}
func (ws *WhileStatement) Statement() {}
func (ws *WhileStatement) String() string {
	return fmt.Sprintf(
		"while (%s) %s",
		ws.Condition,
		ws.Do,
	)
}

type ExpressionStatement struct {
	Expression Expression
}

func (es *ExpressionStatement) Node()      {}
func (es *ExpressionStatement) Statement() {}
func (es *ExpressionStatement) String() string {
	return fmt.Sprintf(
		"%s;",
		es.Expression,
	)
}

type AssignmentStatement struct {
	Left  Expression
	Right Expression
}

func (as *AssignmentStatement) Node()      {}
func (as *AssignmentStatement) Statement() {}
func (as *AssignmentStatement) String() string {
	return fmt.Sprintf(
		"%s = %s;",
		as.Left,
		as.Right,
	)
}

type Declaration struct {
	Identifier *IdentifierLiteral
	Right      Expression
	Mutable    bool
}

func (d *Declaration) Node()      {}
func (d *Declaration) Statement() {}
func (d *Declaration) String() string {
	var modifier string
	if d.Mutable {
		modifier = "var"
	} else {
		modifier = "const"
	}
	return fmt.Sprintf(
		"%s %s = %s;",
		modifier,
		d.Identifier,
		d.Right,
	)
}

type SayStatement struct {
	Expression Expression
}

func (ss *SayStatement) Node()      {}
func (ss *SayStatement) Statement() {}
func (ss *SayStatement) String() string {
	return fmt.Sprintf(
		"say %s;",
		ss.Expression,
	)
}

type ReturnStatement struct {
	Value Expression
}

func (rs *ReturnStatement) Node()      {}
func (rs *ReturnStatement) Statement() {}
func (rs *ReturnStatement) String() string {
	return fmt.Sprintf(
		"return %s;",
		rs.Value,
	)
}

type BreakStatement struct{}

func (bs *BreakStatement) Node()      {}
func (bs *BreakStatement) Statement() {}
func (bs *BreakStatement) String() string {
	return "break;"
}

type ContinueStatement struct{}

func (cs *ContinueStatement) Node()      {}
func (cs *ContinueStatement) Statement() {}
func (cs *ContinueStatement) String() string {
	return "continue;"
}

// == Expression ==

type InfixExpression struct {
	Left     Expression
	Right    Expression
	Operator Operator
}

func (ie *InfixExpression) Node()       {}
func (ie *InfixExpression) Expression() {}
func (ie *InfixExpression) String() string {
	return fmt.Sprintf(
		"(%s %s %s)",
		ie.Left,
		ie.Operator,
		ie.Right,
	)
}

type PrefixExpression struct {
	Right    Expression
	Operator Operator
}

func (pe *PrefixExpression) Node()       {}
func (pe *PrefixExpression) Expression() {}
func (pe *PrefixExpression) String() string {
	return fmt.Sprintf(
		"(%s %s)",
		pe.Operator,
		pe.Right,
	)
}

type CallExpression struct {
	Left      Expression
	Arguments []Expression
}

func (ce *CallExpression) Node()       {}
func (ce *CallExpression) Expression() {}
func (ce *CallExpression) String() string {
	var args strings.Builder
	for i, arg := range ce.Arguments {
		args.WriteString(arg.String())
		if i != len(ce.Arguments)-1 {
			args.WriteString(", ")
		}
	}
	return fmt.Sprintf(
		"%s(%s)",
		ce.Left,
		args.String(),
	)
}

// == Literals ==

type FunctionLiteral struct {
	Body       *Block
	Parameters []*IdentifierLiteral
}

func (fl *FunctionLiteral) Node()       {}
func (fl *FunctionLiteral) Expression() {}
func (fl *FunctionLiteral) String() string {
	var str strings.Builder
	for i, param := range fl.Parameters {
		str.WriteString(param.String())
		if i != len(fl.Parameters)-1 {
			str.WriteString(", ")
		}
	}
	params := str.String()
	return fmt.Sprintf(
		"fun(%s) %s",
		params,
		fl.Body,
	)
}

type NullLiteral struct{}

func (nl *NullLiteral) Node()       {}
func (nl *NullLiteral) Expression() {}
func (nl *NullLiteral) String() string {
	return "null"
}

type BooleanLiteral struct {
	Value bool
}

func (bl *BooleanLiteral) Node()       {}
func (bl *BooleanLiteral) Expression() {}
func (bl *BooleanLiteral) String() string {
	if bl.Value {
		return "true"
	}
	return "false"
}

type NumberLiteral struct {
	Value float64
}

func (nl *NumberLiteral) Node()       {}
func (nl *NumberLiteral) Expression() {}
func (nl *NumberLiteral) String() string {
	return strconv.FormatFloat(nl.Value, 'g', -1, 64)
}

type StringLiteral struct {
	Value string
}

func (sl *StringLiteral) Node()       {}
func (sl *StringLiteral) Expression() {}
func (sl *StringLiteral) String() string {
	return fmt.Sprintf(
		"\"%s\"",
		sl.Value,
	)
}

type IdentifierLiteral struct {
	Value string
}

func (il *IdentifierLiteral) Node()       {}
func (il *IdentifierLiteral) Expression() {}
func (il *IdentifierLiteral) String() string {
	return il.Value
}
