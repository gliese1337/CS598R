package types

import (
	"bytes"
	"fmt"
)

type Evaller func(interface{}, *Environment) interface{}

type Callable interface {
	Call(Evaller, *Environment, *VPair) (interface{}, *Environment, bool)
}

type VSym string

func (v VSym) String() string {
	return string(v)
}

type VBool bool

func (v VBool) String() string {
	if v {
		return "#t"
	}
	return "#f"
}

func (v VBool) Call(eval Evaller, dyn_env *Environment, args *VPair) (interface{}, *Environment, bool) {
	if args == nil {
		return VNil, nil, false
	}
	cdr, ok := args.Cdr.(*VPair)
	if !ok || cdr == nil {
		panic("Invalid Arguments to Branch")
	}
	if bool(v) {
		return args.Car, dyn_env, true
	}
	return cdr.Car, dyn_env, true
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
