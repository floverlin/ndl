package cmd

import (
	"errors"
	"fmt"
	"log"
	"needle/internal/needle/evaluator"
	"needle/internal/needle/lexer"
	"needle/internal/needle/parser"
	"os"
	"strings"
	"time"
)

func RunFile(filePath string) error {
	source, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal("read file: ", err)
	}

	fmt.Println("== Source ==")
	fmt.Println(strings.TrimSpace(string(source)))

	lx := lexer.New(source)

	fmt.Println("== Lexemes ==")
	var lexemes []*lexer.Lexeme
	for {
		lexeme := lx.NextLexeme()
		lexemes = append(lexemes, lexeme)
		if lexeme.Type == lexer.EOF || lexeme.Type == lexer.ERROR {
			break
		}
	}
	lexer.PrintLexemes(lexemes)
	lx.Reset()

	p := parser.New(lx)
	script, err := p.Parse()
	if err != nil {
		fmt.Printf("parse error: %s\n", err)
		return errors.New("compile error")
	}
	fmt.Println("== AST ==")
	fmt.Println(strings.TrimSpace(script.String()))

	glob := evaluator.NewEnv(nil)
	evaluator.LoadBuiltins(glob)
	ev := evaluator.New(glob)
	fmt.Println("== Output ==")
	start := time.Now()
	err = ev.Run(script)
	if err != nil {
		message := fmt.Sprintf("runtime error: %s\n", err)
		fmt.Print(message)
		return errors.New(message)
	}

	fmt.Println("== Results ==")
	fmt.Printf("program ends in %v", time.Since(start))

	return nil
}
