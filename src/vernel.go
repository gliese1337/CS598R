package main

import (
	"os"
	"bufio"
	"fmt"
	"vernel/parser"
	"vernel/eval"
	"vernel/lib"
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

	inchan := make(chan rune)
	go func(){
		freader := bufio.NewReader(file)
		loop: if r, _, err := freader.ReadRune(); err == nil {
			inchan <- r
			goto loop
		}
		close(inchan)
	}()
	for expr := range parser.Parse(inchan) {
		fmt.Printf("Expr: \"%v\"\n",expr)
		fmt.Printf("Value: %v\n", eval.Eval(expr,lib.Standard))
	}
}

