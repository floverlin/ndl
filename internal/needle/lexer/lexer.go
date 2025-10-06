package lexer

import (
	"strings"
)

const eof = 0x00

type Lexer struct {
	source []rune
	arrow  int
	line   int
	column int
}

func New(source []rune) *Lexer {
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
	r := lx.read()

	if (r == '=' || r == '!') && lx.peek() == '=' {
		first := r
		lx.read()
		if lx.peek() == '=' {
			lx.read()
			if first == '=' {
				return NewLexeme(IS, "===", lx.line, lx.column-3)
			} else {
				return NewLexeme(ISNT, "!==", lx.line, lx.column-3)
			}
		} else {
			if first == '=' {
				return NewLexeme(EQ, "==", lx.line, lx.column-2)
			} else {
				return NewLexeme(NE, "!=", lx.line, lx.column-2)
			}
		}
	} else if r == '/' && lx.peek() == '/' {
		lx.read()
		lx.skipComment()
		return lx.NextLexeme()
	} else if r == '/' && lx.peek() == '*' {
		lx.read()
		if errLexeme := lx.skipMultilineComment(); errLexeme != nil {
			return errLexeme
		}
		return lx.NextLexeme()
	} else if t, ok := dual[string([]rune{r, lx.peek()})]; ok {
		literal := string([]rune{r, lx.peek()})
		lx.read()
		return NewLexeme(t, literal, lx.line, lx.column-2)
	} else if t, ok := mono[r]; ok {
		return NewLexeme(t, string(r), lx.line, lx.column-1)
	} else if isAlpha(r) {
		return lx.readIdentifier(r)
	} else if isDigit(r) {
		return lx.readNumber(r)
	} else if r == '"' {
		return lx.readString()
	} else if r == '`' {
		return lx.readUniversalIdentifier()
	} else if r == eof {
		return NewLexeme(EOF, "", lx.line, lx.column-1)
	} else {
		return NewLexeme(ERROR, string(r), lx.line, lx.column-1)
	}
}

func (lx *Lexer) read() rune {
	r := lx.peek()
	if r == '\n' {
		lx.line++
		lx.column = 1
	} else {
		lx.column++
	}
	lx.arrow++
	return r
}

func (lx *Lexer) peek() rune {
	if lx.arrow >= len(lx.source) {
		return eof
	}
	return lx.source[lx.arrow]
}

func (lx *Lexer) skipWhite() {
	for {
		next := lx.peek()
		if next != ' ' &&
			next != '\n' &&
			next != '\t' &&
			next != '\r' {
			return
		}
		lx.read()
	}
}

func (lx *Lexer) skipComment() {
	for {
		r := lx.read()
		if r == '\n' || r == eof {
			return
		}
	}
}

// returns error lexeme if comment is not ended
func (lx *Lexer) skipMultilineComment() *Lexeme {
	line, column := lx.line, lx.column-2
	for {
		r := lx.read()
		if r == '*' && lx.peek() == '/' {
			lx.read()
			return nil
		}
		if r == eof {
			return NewLexeme(ERROR, "/*...", line, column)
		}
	}
}

func (lx *Lexer) readIdentifier(firstChar rune) *Lexeme {
	column := lx.column - 1
	var str strings.Builder
	str.WriteRune(firstChar)
	for {
		next := lx.peek()
		if !isAlpha(next) && !isDigit(next) {
			break
		}
		str.WriteRune(lx.read())
	}
	literal := str.String()
	var type_ LexemeType
	if t, ok := indentifiers[literal]; ok {
		type_ = t
	} else {
		type_ = IDENTIFIER
	}
	return NewLexeme(type_, literal, lx.line, column)
}

func (lx *Lexer) readUniversalIdentifier() *Lexeme {
	if lx.peek() == '`' {
		lx.read()
		return NewLexeme(ERROR, "``", lx.line, lx.column-2)
	}
	column := lx.column - 1
	var str strings.Builder
	for {
		r := lx.read()
		if r == '`' {
			break
		}
		if r == eof || r == '\r' || r == '\n' || r == '\t' {
			return NewLexeme(ERROR, "`"+str.String(), lx.line, column)
		}
		str.WriteRune(r)
	}
	return NewLexeme(IDENTIFIER, str.String(), lx.line, column)
}

func (lx *Lexer) readNumber(firstChar rune) *Lexeme {
	column := lx.column - 1
	var str strings.Builder
	str.WriteRune(firstChar)

	for {
		next := lx.peek()
		if !isDigit(next) && next != '.' {
			break
		}
		if next == '.' {
			str.WriteRune(lx.read())
			for {
				next := lx.peek()
				if !isDigit(next) {
					break
				}
				str.WriteRune(lx.read())
			}
			break
		}
		str.WriteRune(lx.read())
	}
	return NewLexeme(NUMBER, str.String(), lx.line, column)
}

func (lx *Lexer) readString() *Lexeme {
	column := lx.column - 1
	var str strings.Builder
	for {
		r := lx.read()
		if esc, ok := escapes[string([]rune{r, lx.peek()})]; ok {
			str.WriteRune(esc)
			lx.read()
			continue
		}
		if r == '"' {
			break
		}
		if r == '\n' || r == '\r' || r == eof {
			return NewLexeme(ERROR, "\""+str.String(), lx.line, column)
		}
		str.WriteRune(r)
	}
	return NewLexeme(STRING, str.String(), lx.line, column)
}

// Include underscore
func isAlpha(char rune) bool {
	return 'a' <= char && char <= 'z' ||
		'A' <= char && char <= 'Z' ||
		char == '_'
}

func isDigit(char rune) bool {
	return '0' <= char && char <= '9'
}

var mono = map[rune]LexemeType{
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
	'!': WOW,

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

var indentifiers = map[string]LexemeType{
	"or":  OR,
	"and": AND,

	"var": VAR,

	"fun":   FUN,
	"class": CLASS,
	"array": ARRAY,
	"map":   MAP,

	"null":  NULL,
	"true":  BOOLEAN,
	"false": BOOLEAN,

	"for":     FOR,
	"while":   WHILE,
	"do":      DO,
	"if":      IF,
	"else":    ELSE,
	"when":    WHEN,
	"switch":  SWITCH,
	"case":    CASE,
	"default": DEFAULT,
	"throw":   THROW,
	"try":     TRY,
	"catch":   CATCH,
	"finally": FINALLY,

	"this": THIS,

	"return":   RETURN,
	"break":    BREAK,
	"continue": CONTINUE,

	"import": IMPORT,
	"export": EXPORT,

	"say": SAY,
}

var escapes = map[string]rune{
	"\n":   '\n',
	"\r":   '\r',
	"\t":   '\t',
	"\"":   '"',
	"\\\\": '\\',
}
