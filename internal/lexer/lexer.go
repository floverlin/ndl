package lexer

import (
	"strings"
)

const eof = 0x00

type Lexer struct {
	source []byte
	arrow  int
	line   int
	column int
}

func New(source []byte) *Lexer {
	return &Lexer{
		source: source,
		arrow:  0,
		line:   1,
		column: 1,
	}
}

func (lx *Lexer) Reset() {
	lx.arrow = 0
	lx.column = 1
	lx.line = 1
}

func (lx *Lexer) NextLexeme() *Lexeme {
	lx.skipWhite()
	b := lx.read()

	if (b == '=' || b == '!') && lx.current() == '=' {
		first := b
		lx.read()
		if lx.current() == '=' {
			lx.read()
			if first == '=' {
				return &Lexeme{Type: IS, Literal: "===", Line: lx.line, Column: lx.column - 3}
			} else {
				return &Lexeme{Type: ISNT, Literal: "!==", Line: lx.line, Column: lx.column - 3}
			}
		} else {
			if first == '=' {
				return &Lexeme{Type: EQ, Literal: "==", Line: lx.line, Column: lx.column - 2}
			} else {
				return &Lexeme{Type: NE, Literal: "!=", Line: lx.line, Column: lx.column - 2}
			}
		}
	} else if b == '/' && lx.current() == '/' {
		lx.read()
		lx.skipComment()
		return lx.NextLexeme()
	} else if t, ok := dual[string(b)+string(lx.current())]; ok {
		literal := string(b) + string(lx.current())
		lx.read()
		return &Lexeme{
			Type:    t,
			Literal: literal,
			Line:    lx.line,
			Column:  lx.column - 2,
		}
	} else if t, ok := mono[b]; ok {
		return &Lexeme{
			Type:    t,
			Literal: string(b),
			Line:    lx.line,
			Column:  lx.column - 1,
		}
	} else if isAlpha(b) {
		return lx.readIdentifier(b)
	} else if isDigit(b) {
		return lx.readNumber(b)
	} else if b == '"' {
		return lx.readString()
	} else if b == eof {
		return &Lexeme{
			Type:    EOF,
			Literal: "",
			Line:    lx.line,
			Column:  lx.column - 1,
		}
	} else {
		return &Lexeme{
			Type:    ERROR,
			Literal: string(b),
			Line:    lx.line,
			Column:  lx.column - 1,
		}
	}
}

func (lx *Lexer) read() byte {
	b := lx.current()
	if b == '\n' {
		lx.line++
		lx.column = 1
	} else {
		lx.column++
	}
	lx.arrow++
	return b
}

func (lx *Lexer) back() {
	lx.arrow--
	lx.column--
}

func (lx *Lexer) current() byte {
	if lx.arrow >= len(lx.source) {
		return eof
	}
	return lx.source[lx.arrow]
}

func (lx *Lexer) skipWhite() {
	for {
		b := lx.read()
		if b != ' ' &&
			b != '\n' &&
			b != '\t' &&
			b != '\r' {
			lx.back()
			return
		}
	}
}
func (lx *Lexer) skipComment() {
	for {
		b := lx.read()
		if b == '\n' || b == eof {
			return
		}
	}
}
func (lx *Lexer) readIdentifier(firstChar byte) *Lexeme {
	var length int
	var str strings.Builder
	str.WriteByte(firstChar)
	length++
	for {
		b := lx.read()
		if !isAlpha(b) && !isDigit(b) {
			lx.back()
			break
		}
		str.WriteByte(b)
		length++
	}
	literal := str.String()
	var type_ LexemeType
	if t, ok := indentifier[literal]; ok {
		type_ = t
	} else {
		type_ = IDENTIFIER
	}

	return &Lexeme{
		Type:    type_,
		Literal: literal,
		Line:    lx.line,
		Column:  lx.column - length,
	}
}
func (lx *Lexer) readNumber(firstChar byte) *Lexeme {
	var dots, length int
	var str strings.Builder
	str.WriteByte(firstChar)
	length++
	for {
		b := lx.read()
		if !isDigit(b) && b != '.' {
			lx.back()
			break
		}
		if b == '.' {
			dots++
		}
		str.WriteByte(b)
		length++
	}
	literal := str.String()

	if dots > 1 {
		return &Lexeme{
			Type:    ERROR,
			Literal: literal,
			Line:    lx.line,
			Column:  lx.column - length,
		}
	}

	return &Lexeme{
		Type:    NUMBER,
		Literal: literal,
		Line:    lx.line,
		Column:  lx.column - length,
	}
}
func (lx *Lexer) readString() *Lexeme {
	var line, column int = lx.line, lx.column - 1
	var str strings.Builder
	for {
		b := lx.read()
		if b == '"' {
			break
		}
		if b == eof {
			return &Lexeme{
				Type:    ERROR,
				Literal: str.String(),
				Line:    line,
				Column:  column,
			}
		}
		if b == '\n' || b == '\r' {
			return &Lexeme{
				Type:    ERROR,
				Literal: str.String(),
				Line:    line,
				Column:  column,
			}
		}
		str.WriteByte(b)
	}
	literal := str.String()
	return &Lexeme{
		Type:    STRING,
		Literal: literal,
		Line:    line,
		Column:  column,
	}
}

// Include underscore
func isAlpha(char byte) bool {
	return 'a' <= char && char <= 'z' ||
		'A' <= char && char <= 'Z' ||
		char == '_'
}

func isDigit(char byte) bool {
	return '0' <= char && char <= '9'
}

var mono = map[byte]LexemeType{
	'(': L_PAREN,
	')': R_PAREN,
	'{': L_BRACE,
	'}': R_BRACE,
	'[': L_BRACKET,
	']': R_BRACKET,

	';': SEMICOLON,
	':': COLON,
	'?': QUESTION,
	',': COMMA,
	'=': ASSIGN,
	'.': DOT,

	'<': LT,
	'>': GT,

	'+': PLUS,
	'-': MINUS,
	'*': STAR,
	'/': SLASH,
}

var dual = map[string]LexemeType{
	"<=": LE,
	">=": GE,

	"->": ARROW,
}

var indentifier = map[string]LexemeType{
	"or":  OR,
	"and": AND,
	"not": NOT,

	"var":   VAR,
	"const": CONST,
	"fun":   FUN,

	"null":  NULL,
	"true":  BOOLEAN,
	"false": BOOLEAN,

	"for":     FOR,
	"while":   WHILE,
	"do":      DO,
	"if":      IF,
	"else":    ELSE,
	"switch":  SWITCH,
	"case":    CASE,
	"default": DEFAULT,
	"throw":   THROW,
	"try":     TRY,
	"catch":   CATCH,
	"finally": FINALLY,

	"class":       CLASS,
	"constructor": CONSTRUCTOR,
	"get":         GET,
	"set":         SET,
	"private":     PRIVATE,
	"public":      PUBLIC,
	"infix":       INFIX,
	"this":        THIS,

	"return":   RETURN,
	"break":    BREAK,
	"continue": CONTINUE,

	"import": IMPORT,
	"export": EXPORT,

	"say": SAY,
}
