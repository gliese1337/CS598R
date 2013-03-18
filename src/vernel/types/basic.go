package types

import (
	"bytes"
	"fmt"
	"strconv"
)

type VValue interface {
	String() string
}

type Tail struct {
	Expr VValue
	Env  *Environment
	K    *Continuation
}

type Evaller func(*Tail, bool)

type Callable interface {
	VValue
	Call(Evaller, *Tail, *VPair) bool
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

type VBool bool

func (v VBool) String() string {
	if v {
		return "#t"
	}
	return "#f"
}

func (v VBool) Call(eval Evaller, ctx *Tail, args *VPair) bool {
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
	return true
}

type VPair struct {
	Car VValue
	Cdr VValue
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
