package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"vernel/eval"
	"vernel/lib"
	"vernel/parser"
	"vernel/types"
)

func main() {
	var file *os.File
	runtime.GOMAXPROCS(runtime.NumCPU())
	if len(os.Args) > 1 {
		var err error
		file, err = os.Open(os.Args[1])
		if err != nil {
			fmt.Printf("Error opening source file.\n")
			return
		}
		defer file.Close()
	} else {
		file = os.Stdin
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
	env := lib.GetBuiltins()
	reschan := make(chan interface{})
	for expr := range parser.Parse(inchan) {
		go eval.Eval(expr, env, &types.Continuation{
			"Top",
			func(ctx *types.Tail, vals *types.VPair) bool {
				reschan <- vals.Car
				ctx.K = nil
				return false
			},
			nil,
		})
		<-reschan
	}
}
