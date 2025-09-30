package parser

import (
	"fmt"
	"needle/internal/lexer"
	"needle/internal/pkg"
	"strconv"
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
			// case lexer.FUN:
			// 	p.advance()
			// 	if p.match(lexer.IDENTIFIER) {
			// 		p.back()
			// 		return p.declaration()
			// 	}
		}
	}

	switch p.current.Type {
	case lexer.SEMICOLON:
		return newNullStatement(), nil
	case lexer.L_BRACE:
		return p.block()
	case lexer.WHILE:
		return p.whileStatement()
	case lexer.IF:
		return p.ifStatement()
	case lexer.SAY:
		return p.sayStatement()
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
	case lexer.L_PAREN:
		p.advance()
		expr, err = p.expression(LOWEST)
		if err := p.expect(lexer.L_PAREN); err != nil {
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
			lexer.AND, lexer.OR:
			expr, err = p.infixExpression(expr)
		default:
			err = newParseError(
				p.current,
				"unexpected '%s'",
				p.current.Literal,
			)
		}
	}

	return
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

/* == Helpers ================================================================*/

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
	CALL              // . ()
	HIGHEST
)

var precedences = map[lexer.LexemeType]precedence{
	lexer.OR: OR,

	lexer.AND: AND,

	lexer.EQ: EQ,
	lexer.NE: EQ,

	lexer.LT: COMP,
	lexer.LE: COMP,
	lexer.GT: COMP,
	lexer.GE: COMP,

	lexer.PLUS:  TERM,
	lexer.MINUS: TERM,

	lexer.STAR:  FACTOR,
	lexer.SLASH: FACTOR,
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
