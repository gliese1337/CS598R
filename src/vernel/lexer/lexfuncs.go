package lexer

import (
	"bytes"
	"strings"
)

func spaceState(in *runeBuffer, out chan *Item) stateFn {
	for ok := true; ok; _, ok, _ = in.accept(" \t\r\n") {
	}
	return switchState
}

func symbolState(in *runeBuffer, out chan *Item) stateFn {
	buf := new(bytes.Buffer)
	for {
		r, ok, more := in.acceptNot(" \t\r\n()#")
		if !(more && ok) {
			break
		}
		buf.WriteRune(r)
	}
	out <- &Item{Type: "symbol", Token: buf.String()}
	return switchState
}

func numberState(in *runeBuffer, out chan *Item) stateFn {
	buf := new(bytes.Buffer)
	for {
		r, ok, more := in.accept("0123456789.e")
		if !(more && ok) {
			break
		}
		buf.WriteRune(r)
	}
	out <- &Item{Type: "number", Token: buf.String()}
	return switchState
}

func stringState(in *runeBuffer, out chan *Item) stateFn {
	buf := new(bytes.Buffer)
	in.next()
	for {
		r, ok, more := in.acceptNot("\"")
		if !(more && ok) {
			in.next()
			break
		}
		buf.WriteRune(r)
	}
	out <- &Item{Type: "string", Token: buf.String()}
	return switchState
}

func specialState(in *runeBuffer, out chan *Item) stateFn {
	buf := new(bytes.Buffer)
	for {
		r, ok, more := in.acceptNot(" \t\r\n()")
		if !(more && ok) {
			break
		}
		buf.WriteRune(r)
	}
	out <- &Item{Type: "special", Token: buf.String()}
	return switchState
}

func switchState(in *runeBuffer, out chan *Item) stateFn {
	if r, ok := in.peek(); ok {
		switch {
		case strings.IndexRune("().;", r) >= 0:
			in.next()
			out <- &Item{Type: string(r), Token: string(r)}
			return switchState
		case strings.IndexRune("	 \r\n", r) >= 0:
			return spaceState
		case strings.IndexRune("0123456789", r) >= 0:
			return numberState
		case r == '#':
			return specialState
		case r == '"':
			return stringState
		default:
			return symbolState
		}
	}
	out <- &Item{Type: "EOF", Token: "EOF"}
	return nil
}
