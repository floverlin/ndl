package cmd

import (
	"bufio"
	"fmt"
	"needle/internal/interpreter"
	"needle/internal/lexer"
	"needle/internal/parser"
	"os"
)

func RunRepl() error {
	fmt.Println("Needle ver")
	fmt.Println("exit using ctrl+c")
	ev := interpreter.New(interpreter.NewEnv(nil))
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
		_, err = ev.Eval(script)
		if err != nil {
			fmt.Printf("runtime error: %s\n", err)
			continue
		}
	}
}
