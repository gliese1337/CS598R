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

type Evaller func(interface{}, *Environment, *Continuation) interface{}

type Callable interface {
	Call(Evaller, *Environment, *Continuation, *VPair) *Tail
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
func (v VNum) Call(eval Evaller, env *Environment, k *Continuation, args *VPair) *Tail {
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

func (v VBool) Call(eval Evaller, env *Environment, k *Continuation, args *VPair) *Tail {
	if args == nil {
		return k.Fn(&VPair{VNil, VNil})
	}
	cdr, ok := args.Cdr.(*VPair)
	if !ok || cdr == nil {
		panic("Invalid Arguments to Branch")
	}
	if bool(v) {
		return &Tail{args.Car, env, k}
	}
	return &Tail{cdr.Car, env, k}
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
