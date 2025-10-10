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

	lx := lexer.New([]rune(string(source)))

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
	script, errs := p.Parse()
	if errs != nil {
		for _, err := range errs {
			fmt.Printf("parse error: %s\n", err)
		}
		fmt.Println(script)
		return errors.New("error")
	}
	fmt.Println("== AST ==")
	fmt.Println(strings.TrimSpace(script.String()))

	ev := evaluator.New()
	evaluator.LoadBuiltins(ev)
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
