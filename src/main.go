package main

import (
	"fmt"
	"os"
)

func main() {
	buf, err := os.ReadFile("example.alex")
	if err != nil {
		fmt.Printf("error reading file: %v\n", err)
		return
	}
	ts := tokenize(buf)

	p := createParser(ts)
	program := p.parseProgram()

	emitProgram(&program)
}
