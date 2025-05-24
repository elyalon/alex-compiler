package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a file name as argument")
		return
	}
	buf, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("error reading file: %v\n", err)
		return
	}
	ts := tokenize(buf)

	p := createParser(ts)
	program := p.parseProgram()

	emitProgram(&program)
}
