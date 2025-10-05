package cmd

import (
	"bufio"
	"fmt"
	"needle/internal/needle/evaluator"
	"needle/internal/needle/lexer"
	"needle/internal/needle/parser"
	"os"
)

func RunRepl() error {
	fmt.Println("Needle ver")
	fmt.Println("exit using ctrl+c")
	ev := evaluator.New(evaluator.NewEnv(nil))
	for {
		r := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		source, _ := r.ReadBytes('\n')
		source = source[:len(source)-1]
		lx := lexer.New(source)
		script, err := parser.New(lx).Parse()
		if err != nil {
			fmt.Printf("compile error: %s\n", err)
			continue
		}
		if err = ev.Run(script); err != nil {
			fmt.Printf("runtime error: %s\n", err)
			continue
		}
	}
}
