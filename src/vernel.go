package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"vernel/eval"
	"vernel/lib"
	"vernel/parser"
	"vernel/types"
)

var cpuprofile = flag.String("cpu", "", "save cpu profile")
var memprofile = flag.String("mem", "", "save memory profile")
var srcfile = flag.String("f", "", "program file")

func main() {
	var file *os.File
	runtime.MemProfileRate = 1
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			fmt.Printf("Error creating memory log file.\n")
			return
		}
		defer func() {
			pprof.WriteHeapProfile(f)
			f.Close()
		}()
	}
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Printf("Error creating cpu log file.\n")
			return
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if *srcfile != "" {
		var err error
		file, err = os.Open(*srcfile)
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
	for expr := range parser.Parse(inchan) {
		eval.Eval(expr, env, types.Top)
		//	fmt.Printf("%s\n", val)
	}
}
