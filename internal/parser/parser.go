package parser

import (
	"fmt"
	"needle/internal/lexer"
	"needle/internal/pkg"
	"strconv"
)

type Literal string

const (
	LIT_CONSTRUCTOR Literal = "constructor"
	LIT_GET         Literal = "get"
	LIT_SET         Literal = "set"
	LIT_PRIVATE     Literal = "private"
	LIT_PUBLIC      Literal = "public"
	LIT_INFIX       Literal = "infix"
)

type Lexemer interface {
	NextLexeme() *lexer.Lexeme
}

type Parser struct {
	lexemer  Lexemer
	previous *lexer.Lexeme
	current  *lexer.Lexeme
	backpack *pkg.Stack[*lexer.Lexeme]
}

func New(lexemer Lexemer) *Parser {
	p := &Parser{
		lexemer:  lexemer,
		backpack: pkg.NewStack[*lexer.Lexeme](),
	}
	p.advance()
	return p
}

func (p *Parser) Parse() (*Script, error) {
	script := &Script{
		Statements: make([]Statement, 0),
	}

	for !p.match(lexer.EOF) {
		stmt, err := p.statement(true)
		if err != nil {
			return nil, err
		} else {
			script.Statements = append(script.Statements, stmt)
		}
		p.advance()
	}

	return script, nil
}

func (p *Parser) statement(declaration bool) (Statement, error) {
	if declaration {
		switch p.current.Type {
		case lexer.CONST, lexer.VAR:
			return p.declaration()
		case lexer.CLASS:
			p.advance()
			if p.match(lexer.IDENTIFIER) {
				p.back()
				return p.classDeclaration()
			}
			p.back()
		case lexer.FUN:
			p.advance()
			if p.match(lexer.IDENTIFIER) {
				p.back()
				return p.funDeclaration()
			}
			p.back()
		}
	}

	switch p.current.Type {
	case lexer.SEMICOLON:
		return newNullStatement(), nil
	case lexer.L_BRACE:
		return p.block()
	case lexer.WHILE:
		return p.whileStatement()
	case lexer.DO:
		return p.doStatement()
	case lexer.IF:
		return p.ifStatement()
	case lexer.SAY:
		return p.sayStatement()
	case lexer.RETURN:
		return p.returnStatement()
	case lexer.BREAK:
		return p.breakStatement()
	case lexer.CONTINUE:
		return p.continueStatement()
	case lexer.TRY:
		return p.tryStatement()
	case lexer.THROW:
		return p.throwStatement()
	}

	expr, err := p.expression(LOWEST)
	if err != nil {
		return nil, err
	}

	p.advance()
	if p.match(lexer.ASSIGN) {
		return p.assignmentStatement(expr)
	} else if p.match(lexer.SEMICOLON) {
		return &ExpressionStatement{Expression: expr}, nil
	}

	return nil, newParseError(
		p.current,
		"unexpected '%s'",
		p.current.Literal,
	)
}

func (p *Parser) expression(prec precedence) (expr Expression, err error) {
	switch p.current.Type {
	case lexer.CLASS:
		expr, err = p.class()
	case lexer.FUN:
		expr, err = p.fun()
	case lexer.L_PAREN:
		p.advance()
		if p.match(lexer.R_PAREN) {
			return nil, newParseError(
				p.current,
				"unexpected ')'",
			)
		}
		expr, err = p.expression(LOWEST)
		if err := p.expect(lexer.R_PAREN); err != nil {
			return nil, err
		}
	case lexer.NULL:
		expr = &NullLiteral{}
	case lexer.BOOLEAN:
		if val, pErr := strconv.ParseBool(p.current.Literal); pErr != nil {
			err = pErr // TODO
		} else {
			expr = &BooleanLiteral{Value: val}
		}
	case lexer.NUMBER:
		if val, pErr := strconv.ParseFloat(p.current.Literal, 64); pErr != nil {
			err = pErr // TODO
		} else {
			expr = &NumberLiteral{Value: val}
		}
	case lexer.STRING:
		expr = &StringLiteral{Value: p.current.Literal}
	case lexer.IDENTIFIER:
		expr = &IdentifierLiteral{Value: p.current.Literal}
	case lexer.THIS:
		expr = &ThisLiteral{}
	case lexer.MINUS, lexer.PLUS, lexer.NOT:
		op := p.current.Literal
		p.advance()
		if e, pErr := p.expression(UN); pErr != nil {
			err = pErr
		} else {
			expr = &PrefixExpression{Right: e, Operator: op}
		}
	default:
		err = newParseError(
			p.current,
			"unexpected '%s'",
			p.current.Literal,
		)
	}
	if err != nil {
		return
	}

	for prec < p.nextPrecedence() {
		p.advance()
		switch p.current.Type {
		case lexer.PLUS, lexer.MINUS, lexer.STAR, lexer.SLASH,
			lexer.LT, lexer.LE, lexer.GT, lexer.GE, lexer.EQ, lexer.NE,
			lexer.AND, lexer.OR, lexer.IS, lexer.ISNT:
			expr, err = p.infixExpression(expr)
		case lexer.L_PAREN:
			expr, err = p.callExpression(expr)
		case lexer.DOT:
			expr, err = p.propertyExpression(expr)
		default:
			err = newParseError(
				p.current,
				"unexpected '%s'",
				p.current.Literal,
			)
		}
		if err != nil {
			return
		}
	}

	return
}

func (p *Parser) parameters() ([]*IdentifierLiteral, error) {
	params := []*IdentifierLiteral{}
	p.advance()
	if p.match(lexer.R_PAREN) {
		return params, nil
	}
	for {
		if !p.match(lexer.IDENTIFIER) {
			return nil, newParseError(
				p.current,
				"expected 'identifier'",
			)
		}
		params = append(
			params,
			&IdentifierLiteral{Value: p.current.Literal},
		)
		p.advance()
		if p.match(lexer.R_PAREN) {
			break
		}
		if !p.match(lexer.COMMA) {
			return nil, newParseError(
				p.current,
				"expected ',' or ')'",
			)
		}
		p.advance()
		if p.match(lexer.R_PAREN) {
			break
		}
	}
	return params, nil
}

func (p *Parser) arguments() ([]Expression, error) {
	args := []Expression{}
	p.advance()
	if p.match(lexer.R_PAREN) {
		return args, nil
	}
	for {
		expr, err := p.expression(LOWEST)
		if err != nil {
			return nil, err
		}
		args = append(args, expr)
		p.advance()
		if p.match(lexer.R_PAREN) {
			break
		}
		if !p.match(lexer.COMMA) {
			return nil, newParseError(
				p.current,
				"expected ',' or ')'",
			)
		}
		p.advance()
		if p.match(lexer.R_PAREN) {
			break
		}
	}
	return args, nil
}

func (p *Parser) fun() (*FunctionLiteral, error) {
	lit := &FunctionLiteral{}
	if err := p.expect(lexer.L_PAREN); err != nil {
		return nil, err
	}
	params, err := p.parameters()
	if err != nil {
		return nil, err
	}
	lit.Parameters = params
	if err := p.expect(lexer.L_BRACE); err != nil {
		return nil, err
	}
	block, err := p.block()
	if err != nil {
		return nil, err
	}
	lit.Body = block
	return lit, nil
}

func (p *Parser) callExpression(left Expression) (*CallExpression, error) {
	expr := &CallExpression{Left: left}
	args, err := p.arguments()
	if err != nil {
		return nil, err
	}
	expr.Arguments = args
	return expr, nil
}

func (p *Parser) propertyExpression(
	left Expression,
) (*PropertyExpression, error) {
	expr := &PropertyExpression{Left: left}
	if err := p.expect(lexer.IDENTIFIER); err != nil {
		return nil, err
	}
	expr.Property = &IdentifierLiteral{Value: p.current.Literal}
	return expr, nil
}

func (p *Parser) infixExpression(left Expression) (*InfixExpression, error) {
	expr := &InfixExpression{
		Left:     left,
		Operator: p.current.Literal,
	}
	prec := p.currentPrecedence()
	p.advance()
	e, err := p.expression(prec)
	if err != nil {
		return nil, err
	}
	expr.Right = e
	return expr, nil
}

func (p *Parser) assignmentStatement(left Expression) (*AssignmentStatement, error) {
	stmt := &AssignmentStatement{
		Left: left,
	}
	p.advance()
	right, err := p.expression(LOWEST)
	if err != nil {
		return nil, err
	}
	stmt.Right = right
	if err := p.expect(lexer.SEMICOLON); err != nil {
		return nil, err
	}
	return stmt, nil
}

func (p *Parser) declaration() (*Declaration, error) {
	stmt := &Declaration{}
	mutable := p.current.Type == lexer.VAR
	stmt.Mutable = mutable
	if err := p.expect(lexer.IDENTIFIER); err != nil {
		return nil, err
	}
	stmt.Identifier = &IdentifierLiteral{Value: p.current.Literal}
	p.advance()
	if p.match(lexer.SEMICOLON) {
		if !mutable {
			return nil, newParseError(p.current, "expected '='")
		}
		stmt.Right = newNullExpression()
	} else if p.match(lexer.ASSIGN) {
		p.advance()
		right, err := p.expression(LOWEST)
		if err != nil {
			return nil, err
		}
		stmt.Right = right
		p.advance()
	} else {
		if mutable {
			return nil, newParseError(p.current, "expected ';' or '='")
		}
		return nil, newParseError(p.current, "expected '='")
	}
	if !p.match(lexer.SEMICOLON) {
		return nil, newParseError(p.current, "expected ';'")
	}

	return stmt, nil
}

func (p *Parser) funDeclaration() (*Declaration, error) {
	stmt := &Declaration{}
	p.advance()
	stmt.Identifier = &IdentifierLiteral{Value: p.current.Literal}
	fun, err := p.fun()
	if err != nil {
		return nil, err
	}
	stmt.Right = fun
	stmt.Mutable = false
	return stmt, nil
}

func (p *Parser) classDeclaration() (*Declaration, error) {
	stmt := &Declaration{}
	p.advance()
	stmt.Identifier = &IdentifierLiteral{Value: p.current.Literal}
	class, err := p.class()
	if err != nil {
		return nil, err
	}
	stmt.Right = class
	stmt.Mutable = false
	return stmt, nil

}

func (p *Parser) class() (*ClassLiteral, error) {
	lit := &ClassLiteral{
		Constructors: map[*IdentifierLiteral]*FunctionLiteral{},
		Public:       map[*IdentifierLiteral]*FunctionLiteral{},
		Fields:       []*Declaration{},
	}
	if err := p.expect(lexer.L_BRACE); err != nil {
		return nil, err
	}
	p.advance()
	for !p.match(lexer.R_BRACE) {
		if p.current.Type == lexer.VAR || p.current.Type == lexer.CONST {
			decl, err := p.declaration()
			if err != nil {
				return nil, err
			}
			lit.Fields = append(lit.Fields, decl)
		} else if p.matchLiteral(LIT_CONSTRUCTOR) {
			if err := p.expect(lexer.IDENTIFIER); err != nil {
				return nil, err
			}
			name := &IdentifierLiteral{Value: p.current.Literal}
			fun, err := p.fun()
			if err != nil {
				return nil, err
			}
			lit.Constructors[name] = fun
		} else if p.matchLiteral(LIT_PUBLIC) {
			if err := p.expect(lexer.IDENTIFIER); err != nil {
				return nil, err
			}
			name := &IdentifierLiteral{Value: p.current.Literal}
			fun, err := p.fun()
			if err != nil {
				return nil, err
			}
			lit.Public[name] = fun
		}

		p.advance()
		if p.match(lexer.EOF) {
			return nil, newParseError(
				p.current,
				"expected '}'",
			)
		}
	}
	return lit, nil
}

func (p *Parser) block() (*Block, error) {
	block := &Block{Statements: make([]Statement, 0)}
	p.advance()
	for !p.match(lexer.R_BRACE) {
		stmt, err := p.statement(true)
		if err != nil {
			return nil, err
		} else {
			block.Statements = append(block.Statements, stmt)
		}
		p.advance()
		if p.match(lexer.EOF) {
			return nil, newParseError(p.current, "expected '}'")
		}
	}
	return block, nil
}

func (p *Parser) ifStatement() (*IfStatement, error) {
	stmt := &IfStatement{}
	if err := p.expect(lexer.L_PAREN); err != nil {
		return nil, err
	}
	p.advance()
	cond, err := p.expression(LOWEST)
	if err != nil {
		return nil, err
	}
	stmt.Condition = cond
	if err := p.expect(lexer.R_PAREN); err != nil {
		return nil, err
	}
	p.advance()
	then, err := p.statement(false)
	if err != nil {
		return nil, err
	}
	stmt.Then = then
	p.advance()
	if p.match(lexer.ELSE) {
		p.advance()
		else_, err := p.statement(false)
		if err != nil {
			return nil, err
		}
		stmt.Else = else_
	} else {
		p.back()
		stmt.Else = newNullStatement()
	}
	return stmt, nil
}

func (p *Parser) sayStatement() (*SayStatement, error) {
	stmt := &SayStatement{}
	p.advance()
	if p.match(lexer.SEMICOLON) {
		stmt.Expression = newNullExpression()
		return stmt, nil
	}
	expr, err := p.expression(LOWEST)
	if err != nil {
		return nil, err
	}
	stmt.Expression = expr
	if err := p.expect(lexer.SEMICOLON); err != nil {
		return nil, err
	}
	return stmt, nil
}

func (p *Parser) whileStatement() (*WhileStatement, error) {
	stmt := &WhileStatement{}
	if err := p.expect(lexer.L_PAREN); err != nil {
		return nil, err
	}
	p.advance()
	cond, err := p.expression(LOWEST)
	if err != nil {
		return nil, err
	}
	stmt.Condition = cond
	if err := p.expect(lexer.R_PAREN); err != nil {
		return nil, err
	}
	p.advance()
	do, err := p.statement(false)
	if err != nil {
		return nil, err
	}
	stmt.Do = do
	return stmt, nil
}

func (p *Parser) doStatement() (*DoStatement, error) {
	stmt := &DoStatement{}
	p.advance()
	do, err := p.statement(false)
	if err != nil {
		return nil, err
	}
	if err := p.expect(lexer.WHILE); err != nil {
		return nil, err
	}
	if err := p.expect(lexer.L_PAREN); err != nil {
		return nil, err
	}
	p.advance()
	while, err := p.expression(LOWEST)
	if err != nil {
		return nil, err
	}
	if err := p.expect(lexer.R_PAREN); err != nil {
		return nil, err
	}
	if err := p.expect(lexer.SEMICOLON); err != nil {
		return nil, err
	}
	stmt.While = while
	stmt.Do = do
	return stmt, nil
}

func (p *Parser) returnStatement() (*ReturnStatement, error) {
	stmt := &ReturnStatement{}
	p.advance()
	if p.match(lexer.SEMICOLON) {
		stmt.Value = newNullExpression()
		return stmt, nil
	}
	expr, err := p.expression(LOWEST)
	if err != nil {
		return nil, err
	}
	stmt.Value = expr
	if err := p.expect(lexer.SEMICOLON); err != nil {
		return nil, err
	}
	return stmt, nil
}

func (p *Parser) breakStatement() (*BreakStatement, error) {
	stmt := &BreakStatement{}
	if err := p.expect(lexer.SEMICOLON); err != nil {
		return nil, err
	}
	return stmt, nil
}

func (p *Parser) continueStatement() (*ContinueStatement, error) {
	stmt := &ContinueStatement{}
	if err := p.expect(lexer.SEMICOLON); err != nil {
		return nil, err
	}
	return stmt, nil
}

func (p *Parser) tryStatement() (*TryStatement, error) {
	stmt := &TryStatement{}
	p.advance()
	try, err := p.statement(false)
	if err != nil {
		return nil, err
	}
	if err := p.expect(lexer.CATCH); err != nil {
		return nil, err
	}
	if err := p.expect(lexer.L_PAREN); err != nil {
		return nil, err
	}
	if err := p.expect(lexer.IDENTIFIER); err != nil {
		return nil, err
	}
	as := &IdentifierLiteral{Value: p.current.Literal}
	if err := p.expect(lexer.R_PAREN); err != nil {
		return nil, err
	}
	p.advance()
	catch, err := p.statement(false)
	if err != nil {
		return nil, err
	}
	var finally Statement = newNullStatement()
	p.advance()
	if p.match(lexer.FINALLY) {
		p.advance()
		fin, err := p.statement(false)
		if err != nil {
			return nil, err
		}
		finally = fin
	} else {
		p.back()
	}
	stmt.Try = try
	stmt.As = as
	stmt.Catch = catch
	stmt.Finally = finally
	return stmt, nil
}

func (p *Parser) throwStatement() (*ThrowStatement, error) {
	stmt := &ThrowStatement{}
	p.advance()
	errValue, err := p.expression(LOWEST)
	if err != nil {
		return nil, err
	}
	if err := p.expect(lexer.SEMICOLON); err != nil {
		return nil, err
	}
	stmt.Error = errValue
	return stmt, nil
}

/* == Helpers =============================================================== */

func (p *Parser) expect(
	type_ lexer.LexemeType,
	args ...any,
) error {
	p.advance()
	if p.match(type_) {
		return nil
	}
	args = append([]any{type_}, args...)
	return newParseError(p.current, "expected '%s'", args...)
}

func (p *Parser) matchLiteral(literal Literal) bool {
	return p.current.Type == lexer.IDENTIFIER &&
		Literal(p.current.Literal) == literal
}

func (p *Parser) match(reference lexer.LexemeType) bool {
	return p.current.Type == reference
}

func (p *Parser) back() {
	if p.previous == nil {
		panic("double back")
	}
	p.backpack.Push(p.current)
	p.current = p.previous
}

func (p *Parser) advance() {
	p.previous = p.current
	if l, err := p.backpack.Pop(); err == nil {
		p.current = l
		return
	}
	p.current = p.lexemer.NextLexeme()
}

func (p *Parser) currentPrecedence() precedence {
	return precedences[p.current.Type]
}

func (p *Parser) nextPrecedence() precedence {
	p.advance()
	prec := precedences[p.current.Type]
	p.back()
	return prec
}

type precedence uint8

const (
	LOWEST precedence = iota
	OR                // or
	AND               // and
	EQ                // == !=
	COMP              // < <= > >=
	TERM              // + -
	FACTOR            // * /
	UN                // - + not
	CALL              // . () []
	HIGHEST
)

var precedences = map[lexer.LexemeType]precedence{
	lexer.OR: OR,

	lexer.AND: AND,

	lexer.EQ:   EQ,
	lexer.NE:   EQ,
	lexer.IS:   EQ,
	lexer.ISNT: EQ,

	lexer.LT: COMP,
	lexer.LE: COMP,
	lexer.GT: COMP,
	lexer.GE: COMP,

	lexer.PLUS:  TERM,
	lexer.MINUS: TERM,

	lexer.STAR:  FACTOR,
	lexer.SLASH: FACTOR,

	lexer.L_PAREN:   CALL,
	lexer.L_BRACKET: CALL,
	lexer.DOT:       CALL,
}

func newNullStatement() *ExpressionStatement {
	return &ExpressionStatement{
		Expression: &NullLiteral{},
	}
}

func newNullExpression() Expression {
	return &NullLiteral{}
}

func newParseError(at *lexer.Lexeme, message string, args ...any) error {
	message = fmt.Sprintf(message, args...)
	return fmt.Errorf(
		"%s at line %d, column %d",
		message,
		at.Line,
		at.Column,
	)
}
