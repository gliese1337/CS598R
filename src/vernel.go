package main

import (
	"bufio"
	"fmt"
	"os"
	"vernel/eval"
	"vernel/lib"
	"vernel/parser"
)

func main() {
	var file *os.File
	if len(os.Args) < 2 {
		file = os.Stdin
	} else {
		var err error
		file, err = os.Open(os.Args[1])
		if err != nil {
			fmt.Printf("Error opening file.\n")
			return
		}
		defer file.Close()
	}

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}()

	inchan := make(chan rune)
	go func() {
		freader := bufio.NewReader(file)
	loop:
		if r, _, err := freader.ReadRune(); err == nil {
			inchan <- r
			goto loop
		}
		close(inchan)
	}()
	for expr := range parser.Parse(inchan) {
		fmt.Printf("Expr: %s\n", expr)
		val := eval.Eval(expr, lib.Standard)
		fmt.Printf("Value: %s\n", val)
	}
}
