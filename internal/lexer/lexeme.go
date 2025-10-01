package lexer

import (
	"fmt"
	"needle/internal/pkg"
)

type LexemeType string

type Lexeme struct {
	Type    LexemeType
	Literal string
	Line    int
	Column  int
}

func PrintLexemes(lexemes []*Lexeme) {
	fmt.Println("| type         | literal      | line | column |")
	fmt.Println("|--------------|--------------|------|--------|")
	for _, lexeme := range lexemes {
		fmt.Printf(
			"| %-12s | %-12s | %-4d | %-6d |\n",
			lexeme.Type,
			pkg.ShortString(lexeme.Literal, 12),
			lexeme.Line,
			lexeme.Column,
		)
	}
}

const (
	ERROR LexemeType = "__error"
	EOF   LexemeType = "__eof"

	PLUS  LexemeType = "+"
	MINUS LexemeType = "-"
	STAR  LexemeType = "*"
	SLASH LexemeType = "/"

	LT   LexemeType = "<"
	LE   LexemeType = "<="
	GT   LexemeType = ">"
	GE   LexemeType = ">="
	EQ   LexemeType = "=="
	NE   LexemeType = "!="
	IS   LexemeType = "==="
	ISNT LexemeType = "!=="

	ARROW LexemeType = "->"

	L_PAREN   LexemeType = "("
	R_PAREN   LexemeType = ")"
	L_BRACE   LexemeType = "{"
	R_BRACE   LexemeType = "}"
	L_BRACKET LexemeType = "["
	R_BRACKET LexemeType = "]"

	SEMICOLON LexemeType = ";"
	COLON     LexemeType = ":"
	QUESTION  LexemeType = "?"
	COMMA     LexemeType = ","
	ASSIGN    LexemeType = "="
	DOT       LexemeType = "."

	OR  LexemeType = "or"
	AND LexemeType = "and"
	NOT LexemeType = "not"

	VAR   LexemeType = "var"
	CONST LexemeType = "const"
	FUN   LexemeType = "fun"

	IDENTIFIER LexemeType = "identifier"
	NULL       LexemeType = "null"
	BOOLEAN    LexemeType = "boolean"
	NUMBER     LexemeType = "number"
	STRING     LexemeType = "string"

	FOR     LexemeType = "for"
	WHILE   LexemeType = "while"
	DO      LexemeType = "do"
	IF      LexemeType = "if"
	ELSE    LexemeType = "else"
	SWITCH  LexemeType = "switch"
	CASE    LexemeType = "case"
	DEFAULT LexemeType = "default"
	CLASS   LexemeType = "class"
	THROW   LexemeType = "throw"
	TRY     LexemeType = "try"
	CATCH   LexemeType = "catch"
	FINALLY LexemeType = "finally"

	RETURN   LexemeType = "return"
	BREAK    LexemeType = "break"
	CONTINUE LexemeType = "continue"

	IMPORT LexemeType = "import"
	EXPORT LexemeType = "export"

	SAY LexemeType = "say"
)
