package parser

import (
	"fmt"
	"strconv"
	"vernel/lexer"
	. "vernel/types"
)

func parse_special(in string) interface{} {
	switch in {
	case "#t":
		return VBool(true)
	case "#f":
		return VBool(false)
	}
	panic("Invalid Special Form")
	return nil
}

func parse_list(l *lexer.Lexer) (interface{}, bool) {
	token := l.Peek()
	switch token.Type {
	case ")":
		l.Next()
		return VNil, true
	case "EOF":
		return nil, false
	case ".":
		l.Next()
		return parse(l)
	}
	if car, ok := parse(l); ok {
		if cdr, ok := parse_list(l); ok {
			if _, ok := cdr.(*VPair); !ok {
				if l.Next().Type != ")" {
					panic("Can't have multiple cdr expressions")
				}
			}
			return &VPair{car, cdr}, true
		}
	}
	return nil, false
}

func parse(l *lexer.Lexer) (interface{}, bool) {
	token := l.Next()
	switch token.Type {
	case "symbol":
		return VSym(token.Token), true
	case "string":
		return VStr(token.Token), true
	case "number":
		num, err := strconv.ParseFloat(token.Token, 64)
		if err != nil {
			panic(fmt.Sprintf("Invalid Number \"%s\"", token.Token))
		}
		return VNum(num), true
	case "special":
		return parse_special(token.Token), true
	case ";":
		parse(l)
		return parse(l)
	case "EOF":
		return nil, false
	case "(":
		return parse_list(l)
	default:
		panic(fmt.Sprintf("Unexpected Token \"%s\"", token.Token))
	}
	return nil, false
}

func Parse(inc chan rune) chan interface{} {
	outc := make(chan interface{})
	lex := lexer.Lex(inc)
	go (func() {
	loop:
		if obj, ok := parse(lex); ok {
			outc <- obj
			goto loop
		}
		close(outc)
	})()
	return outc
}
