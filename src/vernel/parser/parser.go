package parser
import (
	"vernel/lexer"
	. "vernel/types"
	"fmt"
)

func parse_special(in string) interface{} {
	switch in {
	case "#t": return VBool(true)
	case "#f": return VBool(false)
	}
	panic("Invalid Special Form")
	return nil
}

func parse_list(l *lexer.Lexer) (*VPair, bool) {
	token := l.Peek()
	if token.Type == ")" {
		l.Next()
		return VNil, true
	} else if token.Type == "EOF" {
		return nil, false
	}
	if car, ok := parse(l); ok {
		if cdr, ok := parse_list(l); ok {
			return &VPair{car,cdr}, true
		}
	}
	return nil, false
}

func parse(l *lexer.Lexer) (interface{}, bool) {
	token := l.Next()
	switch {
	case token.Type == "symbol":
		return VSym(token.Token), true
	case token.Type == "special":
		return parse_special(token.Token), true
	case token.Type == "EOF":
		return nil, false
	case token.Type == "(":
		return parse_list(l)
	case token.Type == ")":
		panic("Unexpected )")
	default:
		fmt.Printf("%v\n",token)
		panic("Invalid Token")
	}
	return nil, false
}

func Parse(inc chan rune) chan interface{} {
	outc := make(chan interface{})
	lex := lexer.Lex(inc)
	go (func() {
		loop: if obj, ok := parse(lex); ok {
			outc <- obj
			goto loop
		}
		close(outc)
	})()
	return outc
}
