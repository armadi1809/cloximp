package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) > 1 {
		log.Fatal("Usage: jlox [script]")
	} else if len(args) == 1 {
		runFile(args[0])
	} else {
		repl()
	}
}

func repl() {
	scanner := bufio.NewScanner(os.Stdin)
	vm := &VM{}
	vm.initVM()
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		vm.Interpret(line)
	}
}

func runFile(path string) {
	source, err := os.ReadFile(path)
	if err != nil {
		log.Panicf("An error occurred while reading source file %v", err)
	}
	compiler := &Compiler{}
	compiler.compile(string(source))
}
