package parser

import (
	"needle/internal/needle/lexer"
	"strconv"
)

type Parser2 struct {
	lexemer  Lexemer
	current  *lexer.Lexeme
	backpack *lexer.Lexeme
	errors   []error
}

func New2(lexemer Lexemer) *Parser2 {
	p := &Parser2{
		lexemer:  lexemer,
		backpack: nil,
		errors:   nil,
	}
	p.advance()
	return p
}

func (p *Parser2) Parse() (*Script, []error) {
	script := &Script{
		Statements: make([]Statement, 0),
	}

	for !p.check(lexer.EOF) {
		stmt := p.catch(p.declaration)
		if stmt == nil {
			p.synchronize()
			stmt = newBadStatement()
		}
		script.Statements = append(script.Statements, stmt)
		p.advance()
	}

	return script, p.errors
}

func (p *Parser2) declaration() Statement {
	switch p.current.Type {
	case lexer.VAR:
		return p.varDecl()
	default:
		return p.statement()
	}
}

func (p *Parser2) statement() Statement {
	switch p.current.Type {
	case lexer.SEMICOLON:
		return newNullStatement()
	case lexer.L_BRACE:
		return p.block()
	case lexer.WHILE:
		return p.whileStmt()
	case lexer.DO:
		return p.doStmt()
	case lexer.IF:
		return p.ifStmt()
	case lexer.SAY:
		return p.sayStmt()
	case lexer.TRY:
		return p.tryStmt()
	case lexer.THROW:
		return p.throwtmt()
	case lexer.RETURN:
		return p.returnStmt()
	case lexer.BREAK:
		p.expect(lexer.SEMICOLON)
		return &BreakStatement{}
	case lexer.CONTINUE:
		p.expect(lexer.SEMICOLON)
		return &ContinueStatement{}
	}

	expr := p.expression(LOWEST)
	if p.peek().Type == lexer.ASSIGN {
		p.advance()
		return p.assignStmt(expr)
	}

	p.expect(lexer.SEMICOLON)
	return &ExpressionStatement{Expression: expr}
}

func (p *Parser2) expression(prec precedence) Expression {
	var expr Expression
	switch p.current.Type {
	case lexer.L_PAREN:
		p.advance()
		if p.check(lexer.R_PAREN) {
			panicParseError(
				p.current,
				"unexpected ')'",
			)
		}
		expr = p.expression(LOWEST)
		p.expect(lexer.R_PAREN)

	case lexer.CLASS:
		expr = p.classLit()
	case lexer.FUN:
		expr = p.funLit()

	case lexer.NULL:
		expr = &NullLiteral{}
	case lexer.BOOLEAN:
		if val, err := strconv.ParseBool(p.current.Literal); err != nil {
			panic(err)
		} else {
			expr = &BooleanLiteral{Value: val}
		}
	case lexer.NUMBER:
		if val, err := strconv.ParseFloat(p.current.Literal, 64); err != nil {
			panic(err)
		} else {
			expr = &NumberLiteral{Value: val}
		}
	case lexer.STRING:
		expr = &StringLiteral{Value: p.current.Literal}

	case lexer.IDENTIFIER:
		expr = &IdentifierLiteral{Value: p.current.Literal}
	case lexer.THIS:
		expr = &ThisLiteral{}

	case lexer.MINUS, lexer.PLUS, lexer.WOW:
		op := p.current.Literal
		p.advance()
		e := p.expression(UN)
		expr = &PrefixExpression{Right: e, Operator: op}
	default:
		panicParseError(
			p.current,
			"unexpected '%s'",
			p.current.Literal,
		)
	}

	for prec < p.peekPrecedence() {
		p.advance()
		switch p.current.Type {
		case lexer.PLUS, lexer.MINUS, lexer.STAR, lexer.SLASH,
			lexer.LT, lexer.LE, lexer.GT, lexer.GE, lexer.EQ, lexer.NE,
			lexer.AND, lexer.OR, lexer.IS, lexer.ISNT:
			expr = p.infixExpr(expr)
		case lexer.L_PAREN:
			expr = p.callExpr(expr)
		case lexer.DOT:
			expr = p.propExpr(expr)
		default:
			panicParseError(
				p.current,
				"unexpected '%s'",
				p.current.Literal,
			)
		}
	}

	return expr
}

/* == declarations ===========================================================*/

func (p *Parser2) varDecl() *Declaration {
	stmt := &Declaration{}

	p.expect(lexer.IDENTIFIER)
	stmt.Identifier = &IdentifierLiteral{Value: p.current.Literal}

	p.advance()
	if p.check(lexer.SEMICOLON) {
		stmt.Right = newNullExpression()
		return stmt
	} else if p.check(lexer.ASSIGN) {
		p.advance()
		stmt.Right = p.expression(LOWEST)
		p.expect(lexer.SEMICOLON)
		return stmt
	}
	panicParseError(
		p.current,
		"expected ';' or '='",
	)
	return nil
}

/* == stmt ===================================================================*/

func (p *Parser2) block() *Block {
	block := &Block{
		Statements: make([]Statement, 0),
	}

	p.advance()
	for !p.check(lexer.R_BRACE) {
		stmt := p.catch(p.declaration)
		if stmt == nil {
			p.synchronize()
			stmt = newBadStatement()
		}
		block.Statements = append(block.Statements, stmt)
		p.advance()
		if p.check(lexer.EOF) {
			panicParseError(
				p.current,
				"expected '}'",
			)
		}
	}

	return block
}

func (p *Parser2) whileStmt() *WhileStatement {
	stmt := &WhileStatement{}
	p.expect(lexer.L_PAREN)
	p.advance()
	stmt.Condition = p.expression(LOWEST)
	p.expect(lexer.R_PAREN)
	p.advance()
	stmt.Do = p.statement()
	return stmt
}

func (p *Parser2) doStmt() *DoStatement {
	stmt := &DoStatement{}
	p.advance()
	stmt.Do = p.statement()
	p.expect(lexer.WHILE)
	p.expect(lexer.L_PAREN)
	p.advance()
	stmt.While = p.expression(LOWEST)
	p.expect(lexer.R_PAREN)
	p.expect(lexer.SEMICOLON)
	return stmt
}

func (p *Parser2) ifStmt() *IfStatement {
	stmt := &IfStatement{}
	p.expect(lexer.L_PAREN)
	p.advance()
	stmt.Condition = p.expression(LOWEST)
	p.expect(lexer.R_PAREN)
	p.advance()
	stmt.Then = p.statement()
	if p.peek().Type == lexer.ELSE {
		p.advance()
		p.advance()
		stmt.Else = p.statement()
	} else {
		stmt.Else = newNullStatement()
	}
	return stmt
}

func (p *Parser2) sayStmt() *SayStatement {
	stmt := &SayStatement{}
	p.advance()
	stmt.Expression = p.expression(LOWEST)
	p.expect(lexer.SEMICOLON)
	return stmt
}

func (p *Parser2) tryStmt() *TryStatement {
	stmt := &TryStatement{}
	ended := false
	p.advance()
	stmt.Try = p.statement()
	if p.peek().Type == lexer.CATCH {
		p.advance()
		p.expect(lexer.L_PAREN)
		p.expect(lexer.IDENTIFIER)
		stmt.As = &IdentifierLiteral{Value: p.current.Literal}
		p.expect(lexer.R_PAREN)
		p.advance()
		stmt.Catch = p.statement()
		ended = true
	} else {
		stmt.As = &IdentifierLiteral{Value: "_"}
		stmt.Catch = newNullStatement()
	}
	if p.peek().Type == lexer.FINALLY {
		p.advance()
		p.advance()
		stmt.Finally = p.statement()
		ended = true
	} else {
		stmt.Finally = newNullStatement()
	}
	if !ended {
		panicParseError(
			p.current,
			"expected 'catch' or 'finally'",
		)
	}
	return stmt
}

func (p *Parser2) throwtmt() *ThrowStatement {
	stmt := &ThrowStatement{}
	p.advance()
	stmt.Error = p.expression(LOWEST)
	p.expect(lexer.SEMICOLON)
	return stmt
}

func (p *Parser2) returnStmt() *ReturnStatement {
	stmt := &ReturnStatement{}
	p.advance()
	stmt.Value = p.expression(LOWEST)
	p.expect(lexer.SEMICOLON)
	return stmt
}

func (p *Parser2) assignStmt(left Expression) *AssignmentStatement {
	stmt := &AssignmentStatement{Left: left}
	p.advance()
	stmt.Right = p.expression(LOWEST)
	p.expect(lexer.SEMICOLON)
	return stmt
}

/* == expr ===================================================================*/

func (p *Parser2) classLit() *ClassLiteral {
	lit := &ClassLiteral{
		Constructors: map[*IdentifierLiteral]*FunctionLiteral{},
		Public:       map[*IdentifierLiteral]*FunctionLiteral{},
		Fields:       []*Declaration{},
	}
	p.expect(lexer.L_BRACE)
	p.advance()
	for !p.check(lexer.R_BRACE) {
		if p.current.Type == lexer.VAR {
			decl := p.varDecl()
			lit.Fields = append(lit.Fields, decl)
		} else if p.current.Literal == string(LIT_CONSTRUCTOR) { // TODO
			p.expect(lexer.IDENTIFIER)
			name := &IdentifierLiteral{Value: p.current.Literal}
			lit.Constructors[name] = p.funLit()
		} else if p.current.Literal == string(LIT_PUBLIC) {
			p.expect(lexer.IDENTIFIER)
			name := &IdentifierLiteral{Value: p.current.Literal}
			lit.Public[name] = p.funLit()
		}
		p.advance()
		if p.check(lexer.EOF) {
			panicParseError(
				p.current,
				"expected '}'",
			)
		}
	}
	return lit
}

func (p *Parser2) funLit() *FunctionLiteral {
	lit := &FunctionLiteral{}
	p.expect(lexer.L_PAREN)
	lit.Parameters = p.parameters()
	p.expect(lexer.L_BRACE)
	lit.Body = p.block()
	return lit
}

func (p *Parser2) infixExpr(left Expression) *InfixExpression {
	expr := &InfixExpression{
		Left:     left,
		Operator: p.current.Literal,
	}
	prec := p.currentPrecedence()
	p.advance()
	expr.Right = p.expression(prec)
	return expr
}

func (p *Parser2) callExpr(left Expression) *CallExpression {
	expr := &CallExpression{Left: left}
	expr.Arguments = p.arguments()
	return expr
}

func (p *Parser2) propExpr(left Expression) *PropertyExpression {
	expr := &PropertyExpression{Left: left}
	p.expect(lexer.IDENTIFIER)
	expr.Property = &IdentifierLiteral{Value: p.current.Literal}
	return expr
}

/* == parse utility ==========================================================*/

func (p *Parser2) arguments() []Expression {
	args := []Expression{}
	p.advance()
	if p.check(lexer.R_PAREN) {
		return args
	}
	for {
		expr := p.expression(LOWEST)
		args = append(args, expr)
		p.advance()
		if p.check(lexer.R_PAREN) {
			break
		}
		if !p.check(lexer.COMMA) {
			panicParseError(
				p.current,
				"expected ',' or ')'",
			)
		}
		p.advance()
		if p.check(lexer.R_PAREN) {
			break
		}
	}
	return args
}

func (p *Parser2) parameters() []*IdentifierLiteral {
	params := []*IdentifierLiteral{}
	p.advance()
	if p.check(lexer.R_PAREN) {
		return params
	}
	for {
		if !p.check(lexer.IDENTIFIER) {
			panicParseError(
				p.current,
				"expected 'identifier'",
			)
		}
		params = append(
			params,
			&IdentifierLiteral{Value: p.current.Literal},
		)
		p.advance()
		if p.check(lexer.R_PAREN) {
			break
		}
		if !p.check(lexer.COMMA) {
			panicParseError(
				p.current,
				"expected ',' or ')'",
			)
		}
		p.advance()
		if p.check(lexer.R_PAREN) {
			break
		}
	}
	return params
}

/* == utility =============================================================== */

func (p *Parser2) currentPrecedence() precedence {
	return precedences[p.current.Type]
}

func (p *Parser2) peekPrecedence() precedence {
	return precedences[p.peek().Type]
}

func (p *Parser2) catch(f func() Statement) (result Statement) {
	defer func() {
		if pa := recover(); pa != nil {
			if pErr, ok := pa.(*parseError); ok {
				p.errors = append(p.errors, pErr.Error)
				result = nil
				return
			}
			panic(pa)
		}
	}()
	return f()
}

func (p *Parser2) synchronize() {
	for !p.check(lexer.EOF) {
		if p.check(lexer.SEMICOLON) || p.check(lexer.R_BRACE) {
			return
		}
		switch p.peek().Type {
		case lexer.L_BRACE, lexer.VAR, lexer.WHILE, lexer.DO,
			lexer.SAY, lexer.IF, lexer.RETURN,
			lexer.BREAK, lexer.CONTINUE, lexer.TRY:
			return
		}
		p.advance()
	}
}

// checks next token and stands on it
func (p *Parser2) expect(type_ lexer.LexemeType) {
	p.advance()
	if type_ == p.current.Type {
		return
	}
	panicParseError(p.current, "expected '%s'", type_)
}

// checks current token and jumps on next
func (p *Parser2) consume(type_ lexer.LexemeType) {
	if type_ == p.current.Type {
		p.advance()
		return
	}
	panicParseError(p.current, "expected '%s'", type_)
}

func (p *Parser2) check(t lexer.LexemeType) bool {
	return p.current.Type == t
}

func (p *Parser2) peek() *lexer.Lexeme {
	temp := p.current
	p.advance()
	lxm := p.current
	p.current = temp
	p.backpack = lxm
	return lxm
}

func (p *Parser2) advance() {
	if p.backpack != nil {
		p.current = p.backpack
		p.backpack = nil
		return
	}
	p.current = p.lexemer.NextLexeme()
}
