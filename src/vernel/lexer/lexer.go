package lexer
import (
	"strings"
)

type Item struct {
    Type string
	Token string
}

type stateFn func(*runeBuffer, chan *Item) stateFn

type runeBuffer struct {
	in chan rune
	r rune
	more bool
}

// peek looks at the next rune but doesn't advance the input.
func (rb *runeBuffer) peek() (rune, bool) {
	return rb.r, rb.more
}

// next returns the next rune in the input.
func (rb *runeBuffer) next() (rune, bool) {
	if !rb.more {
		return 0, false
	}
	r := rb.r
	rb.r, rb.more = <- rb.in
	return r, true
}

// accept consumes the next rune if it's from the valid set.
func (rb *runeBuffer) accept(valid string) (rune, bool, bool) {
	r, ok := rb.peek()
	if !ok { return r, false, false }
	if strings.IndexRune(valid, r) >= 0 {
		rb.next()
		return r, true, true
	}
	return r, false, true
}

func (rb *runeBuffer) acceptNot(invalid string) (rune, bool, bool) {
	r, ok := rb.peek()
	if !ok { return r, false, false }
	if strings.IndexRune(invalid, r) < 0 {
		rb.next()
		return r, true, true
	}
	return r, false, true
}

type Lexer struct {
	tokens chan *Item
	next *Item
}

func (l *Lexer) Peek() *Item {
	return l.next
}

func (l *Lexer) Next() *Item {
	item := l.next
	l.next = <- l.tokens
	return item
}

func Lex(in chan rune) *Lexer {
	tokens := make(chan *Item)
	go func(){
		r, more := <- in
		buf := &runeBuffer{
			in:in,
			r:r,
			more:more,
		}
		for state := switchState; state != nil; state = state(buf,tokens) {}
		close(tokens)
	}()
	return &Lexer{
		tokens:tokens,
		next:<-tokens,
	}
}
