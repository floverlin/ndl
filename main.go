package main

import (
	"io"
	"needle/cmd"
	"os"
)

func main() {
	var err error
	if len(os.Args) == 1 {
		err = cmd.RunRepl()
	} else {
		err = cmd.RunFile(os.Args[1])
	}
	if err != nil {
		io.WriteString(os.Stderr, err.Error())
		os.Exit(1)
	}
}
