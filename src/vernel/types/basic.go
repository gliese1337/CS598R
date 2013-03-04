package types

import (
	"bytes"
	"fmt"
	"strconv"
)

type VValue interface {
	Strict(*Tail) bool
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
	Call(Evaller, *Tail, ...VValue) bool
	Arity() (int,bool)
}

type VSym string

func (v VSym) Strict(ctx *Tail) bool {
	ctx.Expr = v
	return false
}
func (v VSym) String() string {
	return string(v)
}

type VStr string

func (v VStr) Strict(ctx *Tail) bool {
	ctx.Expr = v
	return false
}
func (v VStr) String() string {
	return string(v)
}

type VNum float64

func (v VNum) Strict(ctx *Tail) bool {
	ctx.Expr = v
	return false
}
func (v VNum) String() string {
	return strconv.FormatFloat(float64(v), 'g', -1, 64)
}

type VBool bool

func (v VBool) Strict(ctx *Tail) bool {
	ctx.Expr = v
	return false
}
func (v VBool) String() string {
	if v {
		return "#t"
	}
	return "#f"
}
func (v VBool) Arity() (int,bool) {
	return 2, false
}
func (v VBool) Call(eval Evaller, ctx *Tail, args ...VValue) bool {
	if bool(v) {
		ctx.Expr = args[0]
	} else {
		ctx.Expr = args[1]
	}
	return true
}

type VPair struct {
	Car VValue
	Cdr VValue
}

func (v *VPair) Strict(ctx *Tail) bool {
	ctx.Expr = v
	return false
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
