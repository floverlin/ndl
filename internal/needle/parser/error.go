package parser

import (
	"errors"
	"fmt"
	"needle/internal/needle/lexer"
)

type parseError struct {
	Error error
}

func panicParseError(lexeme *lexer.Lexeme, message string, a ...any) {
	finalMessage := fmt.Sprintf(
		"%s at line %d, column %d",
		fmt.Sprintf(message, a...),
		lexeme.Line,
		lexeme.Column,
	)
	panic(&parseError{Error: errors.New(finalMessage)})
}
