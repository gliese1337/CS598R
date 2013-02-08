package types

import (
	"bytes"
	"fmt"
	"strconv"
)

type Tail struct {
	Expr interface{}
	Env  *Environment
	K    *Continuation
}

func (t *Tail) Return(x *VPair) {
	t.K.Fn(t, x)
}

type Evaller func(interface{}, *Environment, *Continuation) interface{}

type Callable interface {
	Call(Evaller, *Tail, *VPair)
}

type VSym string

func (v VSym) String() string {
	return string(v)
}

type VStr string

func (v VStr) String() string {
	return string(v)
}

type VNum float64

func (v VNum) String() string {
	return strconv.FormatFloat(float64(v), 'g', -1, 64)
}

/*
func (v VNum) Call(eval Evaller, ctx *Tail args *VPair) {
	//TODO: Make numbers look like church numerals
	//Short cut- exponentiatiate if the arg is another number
	//Does anything else make sense for non-integers?
}
*/

type VBool bool

func (v VBool) String() string {
	if v {
		return "#t"
	}
	return "#f"
}

func (v VBool) Call(eval Evaller, ctx *Tail, args *VPair) {
	if args == nil {
		ctx.Expr = VNil
	} else {
		cdr, ok := args.Cdr.(*VPair)
		if !ok || cdr == nil {
			panic(fmt.Sprintf("Invalid Arguments to Branch: %v", args))
		}
		if bool(v) {
			ctx.Expr = args.Car
		} else {
			ctx.Expr = cdr.Car
		}
	}
}

type VPair struct {
	Car interface{}
	Cdr interface{}
}

func (v *VPair) String() string {
	if v == nil {
		return "()"
	}
	var buf bytes.Buffer
	buf.WriteRune('(')
write_rest:
	buf.WriteString(fmt.Sprintf("%s", v.Car))
	tail, ok := v.Cdr.(*VPair)
	if ok {
		if tail == nil {
			buf.WriteRune(')')
		} else {
			buf.WriteRune(' ')
			v = tail
			goto write_rest
		}
	} else {
		buf.WriteString(" . ")
		buf.WriteString(fmt.Sprintf("%s", v.Cdr))
		buf.WriteRune(')')
	}
	return buf.String()
}

var VNil *VPair = nil
