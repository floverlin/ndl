package evaluator

import (
	"errors"
	"fmt"
	"maps"
	"needle/internal/needle/parser"
	"needle/internal/pkg"
)

type Evaluator struct {
	env *Env
}

func New(global *Env) *Evaluator {
	return &Evaluator{
		env: global,
	}
}

func (e *Evaluator) Run(script *parser.Script) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case *ContinueSignal:
				err = errors.New("'continue' not inside loop")
			case *BreakSignal:
				err = errors.New("'break' not inside loop")
			case *ReturnSignal:
				err = errors.New("'return' not inside function")
			default:
				panic(r)
			}
		}
	}()
	_, err = e.Eval(script)
	return
}

func (e *Evaluator) Eval(node parser.Node) (Value, error) {
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
		return e.env.Get(node.Value)
	case *parser.ThisLiteral:
		if this := e.env.GetThis(); this == nil {
			return nil, errors.New("'this' is undefined")
		} else {
			return this, nil
		}
	case *parser.NullLiteral:
		return &Null{}, nil
	case *parser.BooleanLiteral:
		return &Boolean{Value: node.Value}, nil
	case *parser.NumberLiteral:
		return &Number{Value: node.Value}, nil
	case *parser.StringLiteral:
		return &String{Value: node.Value}, nil
	case *parser.FunctionLiteral:
		return e.function(node)
	case *parser.ClassLiteral:
		return e.class(node)
	case *parser.ArrayLiteral:
		return e.array(node)
	case *parser.MapLiteral:
		return e.map_(node)
	default:
		return nil, errors.New("TODO")
	}
}

func (e *Evaluator) block(node *parser.Block) (Value, error) {
	oldEnv := e.env
	defer func() { e.env = oldEnv }()
	e.env = NewEnv(oldEnv)
	for _, stmt := range node.Statements {
		_, err := e.Eval(stmt)
		if err != nil {
			return nil, err
		}
	}
	return &Null{}, nil
}

func (e *Evaluator) declaration(node *parser.Declaration) (Value, error) {
	name := node.Identifier.Value
	value, err := e.Eval(node.Right)
	if err != nil {
		return nil, err
	}
	if err := e.env.Declare(name, value); err != nil {
		return nil, err
	}
	return &Null{}, nil
}

func (e *Evaluator) function(node *parser.FunctionLiteral) (Value, error) {
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
	}, nil
}

func (e *Evaluator) class(node *parser.ClassLiteral) (Value, error) {
	class := &Class{}
	fields, err := pkg.SliceToMapMap(
		node.Fields,
		func(decl *parser.Declaration) (string, Value, error) {
			val, err := e.Eval(decl.Right)
			if err != nil {
				return "", nil, err
			}
			return decl.Identifier.Value, val, nil
		},
	)
	if err != nil {
		return nil, err
	}
	f_map_map := func(
		ident *parser.IdentifierLiteral,
		lit *parser.FunctionLiteral,
	) (string, *Function, error) {
		name := ident.Value
		v, err := e.Eval(lit)
		if err != nil {
			return "", nil, err
		}
		return name, v.(*Function), nil
	}
	ctors, err := pkg.MapMap(
		node.Constructors,
		f_map_map,
	)
	if err != nil {
		return nil, err
	}
	public, err := pkg.MapMap(
		node.Public,
		f_map_map,
	)
	if err != nil {
		return nil, err
	}
	private, err := pkg.MapMap(
		node.Private,
		f_map_map,
	)
	if err != nil {
		return nil, err
	}
	getters, err := pkg.MapMap(
		node.Getters,
		f_map_map,
	)
	if err != nil {
		return nil, err
	}
	setters, err := pkg.MapMap(
		node.Setters,
		f_map_map,
	)
	if err != nil {
		return nil, err
	}
	class.Fields = fields
	class.Constructors = ctors
	class.Public = public
	class.Private = private
	class.Getters = getters
	class.Setters = setters
	return class, nil
}

func (e *Evaluator) array(node *parser.ArrayLiteral) (Value, error) {
	arr := &Array{Elements: []Value{}}
	for _, expr := range node.Elements {
		elem, err := e.Eval(expr)
		if err != nil {
			return nil, err
		}
		arr.Elements = append(arr.Elements, elem)
	}
	return arr, nil
}

func (e *Evaluator) map_(node *parser.MapLiteral) (Value, error) {
	m := &Map{Pairs: NewHashTable()}
	for kExpr, vExpr := range node.Pairs {
		k, err := e.Eval(kExpr)
		if err != nil {
			return nil, err
		}
		v, err := e.Eval(vExpr)
		if err != nil {
			return nil, err
		}
		m.Pairs.Set(k, v)
	}
	return m, nil
}

func (e *Evaluator) say(node *parser.SayStatement) (Value, error) {
	value, err := e.Eval(node.Expression)
	if err != nil {
		return nil, err
	}
	fmt.Println(value.Debug())
	return &Null{}, nil
}

func (e *Evaluator) if_(node *parser.IfStatement) (Value, error) {
	cond, err := e.Eval(node.Condition)
	if err != nil {
		return nil, err
	}
	var toDo parser.Node
	if toBoolean(cond) {
		toDo = node.Then
	} else {
		toDo = node.Else
	}
	_, err = e.Eval(toDo)
	if err != nil {
		return nil, err
	}
	return &Null{}, err
}

func (e *Evaluator) while(node *parser.WhileStatement) (Value, error) {
	cond, err := e.Eval(node.Condition)
	if err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(*BreakSignal); ok {
				return
			}
			panic(r)
		}
	}()

	for toBoolean(cond) {
		if err := e.doLoop(node.Do); err != nil {
			return nil, err
		}
		cond, err = e.Eval(node.Condition)
		if err != nil {
			return nil, err
		}
	}
	return &Null{}, nil
}

func (e *Evaluator) do(node *parser.DoStatement) (Value, error) {
	var cond Value = &Boolean{Value: true}
	var err error

	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(*BreakSignal); ok {
				return
			}
			panic(r)
		}
	}()

	for toBoolean(cond) {
		if err := e.doLoop(node.Do); err != nil {
			return nil, err
		}
		cond, err = e.Eval(node.While)
		if err != nil {
			return nil, err
		}
	}
	return &Null{}, nil
}

func (e *Evaluator) doLoop(do parser.Statement) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(*ContinueSignal); ok {
				return
			}
			panic(r)
		}
	}()
	_, err = e.Eval(do)
	return
}

func (e *Evaluator) expression(
	node *parser.ExpressionStatement,
) (Value, error) {
	_, err := e.Eval(node.Expression)
	if err != nil {
		return nil, err
	}
	return &Null{}, err
}

func (e *Evaluator) assignment(
	node *parser.AssignmentStatement,
) (Value, error) {
	right, err := e.Eval(node.Right)
	if err != nil {
		return nil, err
	}
	switch left := node.Left.(type) {
	case *parser.IdentifierLiteral:
		err = e.env.Set(left.Value, right)
	case *parser.PropertyExpression:
		prop := left.Property.Value
		if _, ok := left.Left.(*parser.ThisLiteral); ok {
			instance := e.env.GetThis()
			if instance == nil {
				return nil, errors.New("'this' is undefined")
			}
			if _, ok := instance.(*Instance).Fields[prop]; !ok {
				return nil, errors.New("missing field")
			}
			instance.(*Instance).Fields[prop] = right
			return &Null{}, nil
		}
		obj, err := e.Eval(left.Left)
		if err != nil {
			return nil, err
		}
		switch obj := obj.(type) {
		case *Instance:
			if setter, ok := obj.Class.Setters[prop]; !ok {
				return nil, errors.New("missing setter")
			} else {
				return e.callFunction(setter, obj, []parser.Expression{node.Right}) // TODO
			}
		}
	case *parser.IndexExpression:
		index, err := e.Eval(left.Index)
		if err != nil {
			return nil, err
		}
		obj, err := e.Eval(left.Left)
		if err != nil {
			return nil, err
		}
		switch obj := obj.(type) {
		case *Array:
			index, ok := index.(*Number)
			if !ok {
				return nil, errors.New("non umber index")
			}
			intIndex := int(index.Value)
			if intIndex < 0 || intIndex >= len(obj.Elements) {
				return nil, errors.New("index out of range")
			}
			obj.Elements[intIndex] = right
			return &Null{}, nil
		case *Map:
			_, err := obj.Pairs.Set(index, right)
			if err != nil {
				return nil, err
			}
			return &Null{}, nil
		}
	default:
		err = errors.New("can't assign to ???")
	}
	if err != nil {
		return nil, err
	}
	return &Null{}, err
}

func (e *Evaluator) try(node *parser.TryStatement) (value Value, err error) {
	defer func() {
		if _, err = e.Eval(node.Finally); err != nil {
			value, err = nil, fmt.Errorf("in finally: %w", err)
		}
	}()
	_, err = e.Eval(node.Try)
	if err != nil {
		errValue := &String{Value: err.Error()}
		oldEnv := e.env
		e.env = NewEnv(oldEnv)
		defer func() { e.env = oldEnv }()
		e.env.Declare(node.As.Value, errValue)
		_, err := e.Eval(node.Catch)
		if err != nil {
			return nil, fmt.Errorf("in catch: %w", err)
		}
	}
	return &Null{}, nil
}

func (e *Evaluator) throw(node *parser.ThrowStatement) (Value, error) {
	errValue, err := e.Eval(node.Error)
	if err != nil {
		return nil, err
	}
	return nil, errors.New(errValue.Debug())
}

func (e Evaluator) prefix(node *parser.PrefixExpression) (Value, error) {
	right, err := e.Eval(node.Right)
	if err != nil {
		return nil, err
	}
	if node.Operator == parser.OP_NOT {
		return &Boolean{
			Value: !toBoolean(right),
		}, nil
	}
	if node.Operator == parser.OP_PLUS ||
		node.Operator == parser.OP_MINUS {
		if right.Type() != VAL_NUMBER {
			return nil, fmt.Errorf("expected number, got %s", right.Type())
		}
		if node.Operator == parser.OP_MINUS {
			return &Number{Value: -right.(*Number).Value}, nil
		} else {
			return &Number{Value: +right.(*Number).Value}, nil
		}
	}
	return nil, errors.New("unknown prefix operator")
}

type binOp func(Value, Value) (Value, error)

func (e *Evaluator) infix(node *parser.InfixExpression) (Value, error) {
	left, err := e.Eval(node.Left)
	if err != nil {
		return nil, err
	}
	right, err := e.Eval(node.Right)
	if err != nil {
		return nil, err
	}

	switch node.Operator {
	case parser.OP_IS:
		return &Boolean{Value: right == left}, nil
	case parser.OP_ISNT:
		return &Boolean{Value: right != left}, nil
	case parser.OP_OR:
		if toBoolean(left) {
			return left, nil
		}
		return right, nil
	case parser.OP_AND:
		if toBoolean(left) {
			return right, nil
		}
		return left, nil
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
		return nil, errors.New("unsupported type")
	}
	if !ok {
		return nil, errors.New("unsupported operator for type")
	}
	res, err := f(left, right)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (e *Evaluator) call(
	node *parser.CallExpression,
) (Value, error) {
	left, err := e.Eval(node.Left)
	if err != nil {
		return nil, err
	}
	if fun, ok := left.(*Function); ok {
		return e.callFunction(fun, nil, node.Arguments)
	}
	if method, ok := left.(*Method); ok {
		value, err := e.callFunction(
			method.Function,
			method.This,
			node.Arguments,
		)
		if err != nil {
			return nil, err
		}
		if method.IsConstructor {
			return method.This, nil
		}
		return value, nil
	}
	return nil, errors.New("not collable")
}

func (e *Evaluator) callFunction(
	fun *Function,
	this Value,
	args []parser.Expression,
) (return_ Value, err error) {

	// call native
	if fun.FType == F_NATIVE {
		values, argsErr := e.evalExpressions(args)
		if argsErr != nil {
			return nil, argsErr
		}
		return fun.Native(e, this, values...)
	}

	// call function
	if len(fun.Parameters) != len(args) {
		return nil, fmt.Errorf(
			"expected %d arguments, got %d",
			len(fun.Parameters),
			len(args),
		)
	}
	values, argsErr := e.evalExpressions(args)
	if argsErr != nil {
		return nil, argsErr
	}
	oldEnv := e.env
	defer func() { e.env = oldEnv }()
	e.env = NewEnv(fun.Closure)
	e.env.SetThis(this)
	for i, val := range values {
		e.env.Declare(fun.Parameters[i], val)
	}

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
		_, err := e.Eval(stmt)
		if err != nil {
			return nil, err
		}
	}

	return &Null{}, nil
}

func (e *Evaluator) property(node *parser.PropertyExpression) (Value, error) {
	left, err := e.Eval(node.Left)
	prop := node.Property.Value
	if err != nil {
		return nil, err
	}
	switch left := left.(type) {
	case *Class:
		ctor, ok := left.Constructors[prop]
		if !ok {
			return nil, errors.New("missing constructor")
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
		return method, nil
	case *Instance:
		if _, isThis := node.Left.(*parser.ThisLiteral); isThis {
			value, ok := left.Fields[prop]
			if ok {
				return value, nil
			}
			priv, ok := left.Class.Private[prop]
			if ok {
				method := &Method{
					Function:      priv,
					This:          left,
					IsConstructor: false,
				}
				return method, nil
			}
			pub, ok := left.Class.Public[prop]
			if ok {
				method := &Method{
					Function:      pub,
					This:          left,
					IsConstructor: false,
				}
				return method, nil
			}
			return nil, errors.New("missing field or method")
		}
		if get, ok := left.Class.Getters[prop]; ok {
			val, err := e.callFunction(get, left, []parser.Expression{})
			if err != nil {
				return nil, err
			}
			return val, nil
		}
		if fun, ok := left.Class.Public[prop]; ok {
			method := &Method{
				Function:      fun,
				This:          left,
				IsConstructor: false,
			}
			return method, nil
		}
		return nil, errors.New("missing property")
	default:
		return nil, errors.New("can't get property of ???")
	}
}

func (e *Evaluator) index(node *parser.IndexExpression) (Value, error) {
	left, err := e.Eval(node.Left)
	if err != nil {
		return nil, err
	}
	index, err := e.Eval(node.Index)
	if err != nil {
		return nil, err
	}
	switch left := left.(type) {
	case *Array:
		if index, ok := index.(*Number); !ok {
			return nil, errors.New("non number index")
		} else {
			intIndex := int(index.Value)
			if index.Value != float64(intIndex) {
				return nil, errors.New("non integer index")
			}
			if intIndex < 0 || intIndex >= len(left.Elements) {
				return nil, errors.New("index out of range")
			}
			return left.Elements[intIndex], nil
		}
	case *Map:
		return left.Pairs.Get(index)
	default:
		return nil, errors.New("type not supports index access")
	}
}

func (e *Evaluator) slice(node *parser.SliceExpression) (Value, error) {
	left, err := e.Eval(node.Left)
	if err != nil {
		return nil, err
	}
	start, err := e.Eval(node.Start)
	if err != nil {
		return nil, err
	}
	end, err := e.Eval(node.End)
	if err != nil {
		return nil, err
	}
	switch left := left.(type) {
	case *Array:
		start, startOk := start.(*Number)
		end, endOk := end.(*Number)
		if !startOk || !endOk {
			return nil, errors.New("non number index")
		}
		intStart := int(start.Value)
		if start.Value != float64(intStart) {
			return nil, errors.New("non integer index")
		}
		intEnd := int(end.Value)
		if end.Value != float64(intEnd) {
			return nil, errors.New("non integer index")
		}
		if intStart < 0 || intStart >= len(left.Elements) {
			return nil, errors.New("index out of range")
		}
		if intEnd < 0 || intEnd >= len(left.Elements) {
			return nil, errors.New("index out of range")
		}
		if intStart > intEnd {
			return nil, errors.New("first index greater then second")
		}
		return &Array{Elements: left.Elements[intStart:intEnd]}, nil
	default:
		return nil, errors.New("type not supports index access")
	}
}

func (e *Evaluator) return_(node *parser.ReturnStatement) (Value, error) {
	value, err := e.Eval(node.Value)
	if err != nil {
		return nil, err
	}
	panic(&ReturnSignal{Value: value})
}

func (e *Evaluator) evalExpressions(
	exprs []parser.Expression,
) ([]Value, error) {
	values := []Value{}
	for _, expr := range exprs {
		v, err := e.Eval(expr)
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return values, nil
}

func (e *Evaluator) script(node *parser.Script) (*Null, error) {
	for _, stmt := range node.Statements {
		_, err := e.Eval(stmt)
		if err != nil {
			return nil, err
		}
	}
	return &Null{}, nil
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
