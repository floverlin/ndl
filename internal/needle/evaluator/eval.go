package evaluator

import (
	"errors"
	"fmt"
	"maps"
	"needle/internal/needle/parser"
	"needle/internal/pkg"
)

type Evaluator struct {
	env            *Env
	callStack      *pkg.Stack[*Function]
	defaultClasses map[string]*Class
}

func New() *Evaluator {
	env := NewEnv(nil)
	classes := CreateBaseClasses()
	for name, class := range classes {
		env.Declare(name, class)
	}
	return &Evaluator{
		env:            env,
		callStack:      pkg.NewStack[*Function](),
		defaultClasses: classes,
	}
}

func (e *Evaluator) Run(script *parser.Script) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if exc, ok := r.(*Exception); ok {
				err = exc
				return
			}
			panic(r)
		}
	}()
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case *ContinueSignal:
				e.ThrowException("'continue' outside loop")
			case *BreakSignal:
				e.ThrowException("'break' outside loop or switch")
			case *ReturnSignal:
				e.ThrowException("'return' outside function")
			default:
				panic(r)
			}
		}
	}()
	e.Eval(script)
	return nil
}

func (e *Evaluator) Eval(node parser.Node) Value {
	switch node := node.(type) {
	case *parser.Script:
		return e.script(node)
	case *parser.Block:
		return e.block(node)
	case *parser.Declaration:
		return e.declaration(node)
	case *parser.SayStatement:
		return e.say(node)
	case *parser.IfStatement:
		return e.if_(node)
	case *parser.WhileStatement:
		return e.while(node)
	case *parser.DoStatement:
		return e.do(node)
	case *parser.ExpressionStatement:
		return e.expression(node)
	case *parser.AssignmentStatement:
		return e.assignment(node)
	case *parser.ReturnStatement:
		return e.return_(node)
	case *parser.BreakStatement:
		panic(&BreakSignal{})
	case *parser.ContinueStatement:
		panic(&ContinueSignal{})
	case *parser.TryStatement:
		return e.try(node)
	case *parser.ThrowStatement:
		return e.throw(node)

	case *parser.InfixExpression:
		return e.infix(node)
	case *parser.PrefixExpression:
		return e.prefix(node)
	case *parser.CallExpression:
		return e.call(node)
	case *parser.PropertyExpression:
		return e.property(node)
	case *parser.IndexExpression:
		return e.index(node)
	case *parser.SliceExpression:
		return e.slice(node)

	case *parser.IdentifierLiteral:
		val, err := e.env.Get(node.Value)
		if err != nil {
			e.ThrowException("%s", err.Error())
		}
		return val
	case *parser.ThisLiteral:
		if this := e.env.GetThis(); this == nil {
			e.ThrowException("'this' is undefined")
		} else {
			return this
		}
	case *parser.NullLiteral:
		return e.env.globals.Null
	case *parser.BooleanLiteral:
		if node.Value {
			return e.env.globals.True
		}
		return e.env.globals.False
	case *parser.NumberLiteral:
		return &Number{Value: node.Value}
	case *parser.StringLiteral:
		return &String{Value: node.Value}
	case *parser.FunctionLiteral:
		return e.function(node)
	case *parser.ClassLiteral:
		return e.class(node)
	case *parser.ArrayLiteral:
		return e.array(node)
	case *parser.TableLiteral:
		return e.table(node)
	default:
		e.ThrowException("TODO")
	}
	return nil
}

func (e *Evaluator) block(node *parser.Block) Value {
	oldEnv := e.env
	defer func() { e.env = oldEnv }()
	e.env = NewEnv(oldEnv)
	for _, stmt := range node.Statements {
		e.Eval(stmt)
	}
	return nil
}

func (e *Evaluator) declaration(node *parser.Declaration) Value {
	name := node.Identifier.Value
	if err := e.env.Declare(name, e.Eval(node.Right)); err != nil {
		e.ThrowException("%s", err.Error())
	}
	return nil
}

func (e *Evaluator) function(node *parser.FunctionLiteral) Value {
	params, _ := pkg.SliceMap(
		node.Parameters,
		func(e *parser.IdentifierLiteral) (string, error) {
			return e.Value, nil
		},
	)
	return &Function{
		FType:      F_FUNCTION,
		Closure:    e.env.Clone(),
		Body:       node.Body.Statements,
		Parameters: params,
	}
}

func (e *Evaluator) class(node *parser.ClassLiteral) Value {
	class := &Class{}
	fields, _ := pkg.SliceToMapMap(
		node.Fields,
		func(decl *parser.Declaration) (string, Value, error) {
			return decl.Identifier.Value, e.Eval(decl.Right), nil
		},
	)
	f_map_map := func(
		ident *parser.IdentifierLiteral,
		lit *parser.FunctionLiteral,
	) (string, *Function, error) {
		return ident.Value, e.Eval(lit).(*Function), nil
	}
	ctors, _ := pkg.MapMap(
		node.Constructors,
		f_map_map,
	)
	public, _ := pkg.MapMap(
		node.Public,
		f_map_map,
	)
	private, _ := pkg.MapMap(
		node.Private,
		f_map_map,
	)
	getters, _ := pkg.MapMap(
		node.Getters,
		f_map_map,
	)
	setters, _ := pkg.MapMap(
		node.Setters,
		f_map_map,
	)
	class.Fields = fields
	class.Constructors = ctors
	class.Public = public
	class.Private = private
	class.Getters = getters
	class.Setters = setters
	return class
}

func (e *Evaluator) array(node *parser.ArrayLiteral) Value {
	arr := &Array{Elements: []Value{}}
	for _, expr := range node.Elements {
		arr.Elements = append(arr.Elements, e.Eval(expr))
	}
	return arr
}

func (e *Evaluator) table(node *parser.TableLiteral) Value {
	table := &Table{Pairs: NewHashTable()}
	for kExpr, vExpr := range node.Pairs {
		table.Pairs.Set(e.Eval(kExpr), e.Eval(vExpr))
	}
	return table
}

func (e *Evaluator) say(node *parser.SayStatement) Value {
	fmt.Println(e.Eval(node.Expression).Debug())
	return nil
}

func (e *Evaluator) if_(node *parser.IfStatement) Value {
	var toDo parser.Node
	if toBoolean(e.Eval(node.Condition)) {
		toDo = node.Then
	} else {
		toDo = node.Else
	}
	e.Eval(toDo)
	return nil
}

func (e *Evaluator) while(node *parser.WhileStatement) Value {
	cond := e.Eval(node.Condition)

	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(*BreakSignal); ok {
				return
			}
			panic(r)
		}
	}()

	for toBoolean(cond) {
		e.loop(node.Do)
		cond = e.Eval(node.Condition)
	}
	return nil
}

func (e *Evaluator) do(node *parser.DoStatement) Value {
	var cond Value = &Boolean{Value: true}

	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(*BreakSignal); ok {
				return
			}
			panic(r)
		}
	}()

	for toBoolean(cond) {
		e.loop(node.Do)
		cond = e.Eval(node.While)
	}
	return nil
}

func (e *Evaluator) loop(do parser.Statement) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(*ContinueSignal); ok {
				return
			}
			panic(r)
		}
	}()
	e.Eval(do)
}

func (e *Evaluator) expression(
	node *parser.ExpressionStatement,
) Value {
	e.Eval(node.Expression)
	return nil
}

func (e *Evaluator) assignment(
	node *parser.AssignmentStatement,
) Value {
	right := e.Eval(node.Right)

	switch left := node.Left.(type) {
	case *parser.IdentifierLiteral:
		if err := e.env.Set(left.Value, right); err != nil {
			e.ThrowException("%s", err.Error())
		}
	case *parser.PropertyExpression:
		prop := left.Property.Value
		if _, isThis := left.Left.(*parser.ThisLiteral); isThis {
			instance := e.env.GetThis()
			if instance == nil {
				e.ThrowException("'this' is undefined")
			}
			if _, ok := instance.(*Instance).Fields[prop]; !ok {
				e.ThrowException("missing field")
			}
			instance.(*Instance).Fields[prop] = right
			return nil
		}
		obj := e.Eval(left.Left)
		switch obj := obj.(type) {
		case *Instance:
			if setter, ok := obj.Class.Setters[prop]; !ok {
				e.ThrowException("missing setter")
			} else {
				e.callFunction(setter, obj, []parser.Expression{node.Right}) // TODO
				return nil
			}
		}
	case *parser.IndexExpression:
		index := e.Eval(left.Index)
		obj := e.Eval(left.Left)
		switch obj := obj.(type) {
		case *Array:
			index, ok := index.(*Number)
			if !ok {
				e.ThrowException("non umber index")
			}
			intIndex := int(index.Value)
			if intIndex < 0 || intIndex >= len(obj.Elements) {
				e.ThrowException("index out of range")
			}
			obj.Elements[intIndex] = right
			return nil
		case *Table:
			_, err := obj.Pairs.Set(index, right)
			if err != nil {
				e.ThrowException("%s", err.Error())
			}
			return nil
		}
	default:
		e.ThrowException("can't assign to ???")
	}
	return nil
}

func (e *Evaluator) try(node *parser.TryStatement) Value {
	_, exc := pkg.Catch[parser.Node, Value, *Exception](e.Eval, node.Try)
	var excCatch *Exception
	if exc != nil {
		oldEnv := e.env
		e.env = NewEnv(oldEnv)
		defer func() { e.env = oldEnv }()
		e.env.Declare(node.As.Value, exc)
		_, excCatch = pkg.Catch[parser.Node, Value, *Exception](e.Eval, node.Catch)
	}
	_, excFin := pkg.Catch[parser.Node, Value, *Exception](e.Eval, node.Finally)

	if excFin != nil {
		panic(excFin)
	} else if excCatch != nil {
		panic(excCatch)
	}
	return nil
}

func (e *Evaluator) throw(node *parser.ThrowStatement) Value {
	e.ThrowException("%s", e.Eval(node.Error).Debug())
	return nil
}

func (e Evaluator) prefix(node *parser.PrefixExpression) Value {
	right := e.Eval(node.Right)
	if node.Operator == parser.OP_NOT {
		return &Boolean{
			Value: !toBoolean(right),
		}
	}
	if node.Operator == parser.OP_PLUS ||
		node.Operator == parser.OP_MINUS {
		if right.Type() != VAL_NUMBER {
			e.ThrowException("expected number, got %s", right.Type())
		}
		if node.Operator == parser.OP_MINUS {
			return &Number{Value: -right.(*Number).Value}
		} else {
			return &Number{Value: +right.(*Number).Value}
		}
	}
	e.ThrowException("unknown prefix operator")
	return nil
}

type binOp func(Value, Value) (Value, error)

func (e *Evaluator) infix(node *parser.InfixExpression) Value {
	left := e.Eval(node.Left)
	right := e.Eval(node.Right)

	switch node.Operator {
	case parser.OP_IS:
		return &Boolean{Value: right == left}
	case parser.OP_ISNT:
		return &Boolean{Value: right != left}
	case parser.OP_OR:
		if toBoolean(left) {
			return left
		}
		return right
	case parser.OP_AND:
		if toBoolean(left) {
			return right
		}
		return left
	}

	var f binOp
	var ok bool
	switch left.(type) {
	case *Number:
		f, ok = numBinOps[node.Operator]
	case *String:
		f, ok = strBinOps[node.Operator]
	case *Boolean:
		f, ok = boolBinOps[node.Operator]
	default:
		e.ThrowException("unsupported type")
	}
	if !ok {
		e.ThrowException("unsupported operator for type")
	}
	res, err := f(left, right)
	if err != nil {
		e.ThrowException("%s", err.Error())
	}
	return res
}

func (e *Evaluator) call(
	node *parser.CallExpression,
) Value {
	left := e.Eval(node.Left)
	if fun, ok := left.(*Function); ok {
		return e.callFunction(fun, nil, node.Arguments)
	}
	if method, ok := left.(*Method); ok {
		value := e.callFunction(
			method.Function,
			method.This,
			node.Arguments,
		)
		if method.IsConstructor {
			return method.This
		}
		return value
	}
	e.ThrowException("not collable")
	return nil
}

func (e *Evaluator) callFunction(
	fun *Function,
	this Value,
	args []parser.Expression,
) (return_ Value) {
	catchSignal := func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case *ContinueSignal:
				e.ThrowException("'continue' outside loop")
			case *BreakSignal:
				e.ThrowException("'break' outside loop or switch")
			default:
				panic(r)
			}
		}
	}

	// call native
	if fun.FType == F_NATIVE {
		values := e.evalExpressions(args)

		e.callStack.Push(fun)
		defer e.callStack.Pop()

		defer catchSignal()
		return fun.Native(e, this, values...)
	}

	// call function
	if len(fun.Parameters) != len(args) {
		e.ThrowException(
			"expected %d arguments, got %d",
			len(fun.Parameters),
			len(args),
		)
	}
	values := e.evalExpressions(args)

	oldEnv := e.env
	defer func() { e.env = oldEnv }()
	e.env = NewEnv(fun.Closure)
	e.env.SetThis(this)
	for i, val := range values {
		e.env.Declare(fun.Parameters[i], val)
	}

	e.callStack.Push(fun)
	defer e.callStack.Pop()

	defer catchSignal()

	defer func() {
		if r := recover(); r != nil {
			if rs, ok := r.(*ReturnSignal); ok {
				return_ = rs.Value
				return
			}
			panic(r)
		}
	}()

	for _, stmt := range fun.Body {
		e.Eval(stmt)
	}

	return e.env.globals.Null
}

func (e *Evaluator) property(node *parser.PropertyExpression) Value {
	left := e.Eval(node.Left)
	prop := node.Property.Value

	switch left := left.(type) {
	case *Class:
		ctor, ok := left.Constructors[prop]
		if !ok {
			e.ThrowException("missing constructor")
		}
		this := &Instance{
			Class:  left,
			Fields: maps.Clone(left.Fields),
		}
		method := &Method{
			Function:      ctor,
			This:          this,
			IsConstructor: true,
		}
		return method
	case *Instance:
		if _, isThis := node.Left.(*parser.ThisLiteral); isThis {
			value, ok := left.Fields[prop]
			if ok {
				return value
			}
			priv, ok := left.Class.Private[prop]
			if ok {
				method := &Method{
					Function:      priv,
					This:          left,
					IsConstructor: false,
				}
				return method
			}
			pub, ok := left.Class.Public[prop]
			if ok {
				return &Method{
					Function:      pub,
					This:          left,
					IsConstructor: false,
				}
			}
			e.ThrowException("missing field or method")
		}
		if get, ok := left.Class.Getters[prop]; ok {
			return e.callFunction(get, left, []parser.Expression{})
		}
		if fun, ok := left.Class.Public[prop]; ok {
			return &Method{
				Function:      fun,
				This:          left,
				IsConstructor: false,
			}
		}
		e.ThrowException("missing property")
	case *String:
		pub, ok := e.defaultClasses[CLASS_STRING].Public[prop]
		if ok {
			return &Method{
				Function:      pub,
				This:          left,
				IsConstructor: false,
			}
		}
		e.ThrowException("missing field or method")
	case *Number:
		pub, ok := e.defaultClasses[CLASS_NUMBER].Public[prop]
		if ok {
			return &Method{
				Function:      pub,
				This:          left,
				IsConstructor: false,
			}
		}
		e.ThrowException("missing field or method")
	case *Array:
		pub, ok := e.defaultClasses[CLASS_ARRAY].Public[prop]
		if ok {
			return &Method{
				Function:      pub,
				This:          left,
				IsConstructor: false,
			}
		}
		e.ThrowException("missing field or method")
	case *Table:
		pub, ok := e.defaultClasses[CLASS_TABLE].Public[prop]
		if ok {
			return &Method{
				Function:      pub,
				This:          left,
				IsConstructor: false,
			}
		}
		e.ThrowException("missing field or method")
	}
	e.ThrowException("can't get property of ???")
	return nil
}

func (e *Evaluator) index(node *parser.IndexExpression) Value {
	left := e.Eval(node.Left)
	index := e.Eval(node.Index)

	switch left := left.(type) {
	case *Array:
		if index, ok := index.(*Number); !ok {
			e.ThrowException("non number index")
		} else {
			intIndex := int(index.Value)
			if index.Value != float64(intIndex) {
				e.ThrowException("non integer index")
			}
			if intIndex < 0 || intIndex >= len(left.Elements) {
				e.ThrowException("index out of range")
			}
			return left.Elements[intIndex]
		}
	case *Table:
		val, err := left.Pairs.Get(index)
		if err != nil {
			e.ThrowException("%s", err.Error())
		}
		return val
	}
	e.ThrowException("type not supports index access")
	return nil
}

func (e *Evaluator) slice(node *parser.SliceExpression) Value {
	left := e.Eval(node.Left)
	start := e.Eval(node.Start)
	end := e.Eval(node.End)

	switch left := left.(type) {
	case *Array:
		start, startOk := start.(*Number)
		end, endOk := end.(*Number)
		if !startOk || !endOk {
			e.ThrowException("non number index")
		}
		intStart := int(start.Value)
		if start.Value != float64(intStart) {
			e.ThrowException("non integer index")
		}
		intEnd := int(end.Value)
		if end.Value != float64(intEnd) {
			e.ThrowException("non integer index")
		}
		if intStart < 0 || intStart >= len(left.Elements) {
			e.ThrowException("index out of range")
		}
		if intEnd < 0 || intEnd >= len(left.Elements) {
			e.ThrowException("index out of range")
		}
		if intStart > intEnd {
			e.ThrowException("first index greater then second")
		}
		return &Array{Elements: left.Elements[intStart:intEnd]}
	}
	e.ThrowException("type not supports index access")
	return nil
}

func (e *Evaluator) return_(node *parser.ReturnStatement) Value {
	panic(&ReturnSignal{Value: e.Eval(node.Value)})
}

func (e *Evaluator) evalExpressions(
	exprs []parser.Expression,
) []Value {
	values := []Value{}
	for _, expr := range exprs {
		values = append(values, e.Eval(expr))
	}
	return values
}

func (e *Evaluator) script(node *parser.Script) Value {
	for _, stmt := range node.Statements {
		e.Eval(stmt)
	}
	return nil
}

func toBoolean(value Value) bool {
	if value.Type() == VAL_NULL {
		return false
	}
	if value.Type() == VAL_BOOLEAN {
		return value.(*Boolean).Value
	}
	return true
}

/* == bin ops ================================================================*/

var boolBinOps = map[parser.Operator]binOp{
	parser.OP_EQ: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_BOOLEAN {
			return &Boolean{Value: false}, nil
		}
		return &Boolean{Value: v1.(*Boolean).Value == v2.(*Boolean).Value}, nil
	},
	parser.OP_NE: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_BOOLEAN {
			return &Boolean{Value: true}, nil
		}
		return &Boolean{Value: v1.(*Boolean).Value != v2.(*Boolean).Value}, nil
	},
}

var strBinOps = map[parser.Operator]binOp{
	parser.OP_PLUS: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_STRING {
			return nil, errors.New("expected string")
		}
		return &String{Value: v1.(*String).Value + v2.(*String).Value}, nil
	},
	parser.OP_EQ: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_STRING {
			return &Boolean{Value: false}, nil
		}
		return &Boolean{Value: v1.(*String).Value == v2.(*String).Value}, nil
	},
	parser.OP_NE: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_STRING {
			return &Boolean{Value: true}, nil
		}
		return &Boolean{Value: v1.(*String).Value != v2.(*String).Value}, nil
	},
}

var numBinOps = map[parser.Operator]binOp{
	parser.OP_PLUS: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_NUMBER {
			return nil, errors.New("expected number")
		}
		return &Number{Value: v1.(*Number).Value + v2.(*Number).Value}, nil
	},
	parser.OP_MINUS: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_NUMBER {
			return nil, errors.New("expected number")
		}
		return &Number{Value: v1.(*Number).Value - v2.(*Number).Value}, nil
	},
	parser.OP_STAR: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_NUMBER {
			return nil, errors.New("expected number")
		}
		return &Number{Value: v1.(*Number).Value * v2.(*Number).Value}, nil
	},
	parser.OP_SLASH: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_NUMBER {
			return nil, errors.New("expected number")
		}
		return &Number{Value: v1.(*Number).Value / v2.(*Number).Value}, nil
	},
	parser.OP_EQ: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_NUMBER {
			return &Boolean{Value: false}, nil
		}
		return &Boolean{Value: v1.(*Number).Value == v2.(*Number).Value}, nil
	},
	parser.OP_LT: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_NUMBER {
			return nil, errors.New("expected number")
		}
		return &Boolean{Value: v1.(*Number).Value < v2.(*Number).Value}, nil
	},
	parser.OP_LE: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_NUMBER {
			return nil, errors.New("expected number")
		}
		return &Boolean{Value: v1.(*Number).Value <= v2.(*Number).Value}, nil
	},
	parser.OP_NE: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_NUMBER {
			return &Boolean{Value: true}, nil
		}
		return &Boolean{Value: v1.(*Number).Value != v2.(*Number).Value}, nil
	},
	parser.OP_GT: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_NUMBER {
			return nil, errors.New("expected number")
		}
		return &Boolean{Value: v1.(*Number).Value > v2.(*Number).Value}, nil
	},
	parser.OP_GE: func(v1, v2 Value) (Value, error) {
		if v2.Type() != VAL_NUMBER {
			return nil, errors.New("expected number")
		}
		return &Boolean{Value: v1.(*Number).Value >= v2.(*Number).Value}, nil
	},
}

func (e *Evaluator) ThrowException(message string, a ...any) {
	panic(&Exception{
		Message:    fmt.Sprintf(message, a...),
		StackTrace: e.callStack.Shot(),
	})
}
