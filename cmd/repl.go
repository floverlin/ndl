package cmd

import (
	"bufio"
	"fmt"
	"needle/internal/needle"
	"os"
)

func RunRepl() error {
	fmt.Println("Needle ver0.0.1")
	fmt.Println("exit using ctrl+c")
	n := needle.New()
	needle.LoadBuiltin(n)
	for {
		fmt.Print("> ")
		r := bufio.NewReader(os.Stdin)
		str, _ := r.ReadString('\n')
		err := n.RunString(str[:len(str)-2])
		if err != nil {
			fmt.Println(err)
		}
	}
}
