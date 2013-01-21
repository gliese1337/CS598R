package lexer
import (
	"bytes"
	"strings"
)

func spaceState(in *runeBuffer, out chan *Item) stateFn {
	for ok := true; ok; _, ok, _ = in.accept(" \t\r\n") {}
	return switchState
}

func symbolState(in *runeBuffer, out chan *Item) stateFn {
	buf := new(bytes.Buffer)
	for {
		r, ok, more := in.acceptNot(" \t\r\n()#")
		if !(more && ok) { break }
		buf.WriteRune(r)
	}
	out <- &Item{Type:"symbol",Token:buf.String()}
	return switchState
}

func specialState(in *runeBuffer, out chan *Item) stateFn {
	buf := new(bytes.Buffer)
	for {
		r, ok, more := in.acceptNot(" \t\r\n()")
		if !(more && ok) { break }
		buf.WriteRune(r)
	}
	out <- &Item{Type:"special",Token:buf.String()}
	return switchState
}

func switchState(in *runeBuffer, out chan *Item) stateFn {
	if r, ok := in.peek(); ok {
		switch {
		case strings.IndexRune("	 \r\n", r) >= 0:
			return spaceState
		case r == '(':
			in.next()
			out <- &Item{Type:"(",Token:"("}
			return switchState
		case r == ')':
			in.next()
			out <- &Item{Type:")",Token:")"}
			return switchState
		case r == '#':
			return specialState
		default:
			return symbolState
		}
	}
	out <- &Item{Type:"EOF",Token:"EOF"}
	return nil
}
