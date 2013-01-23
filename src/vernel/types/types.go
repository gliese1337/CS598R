package types

import (
	"bytes"
	"fmt"
)

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

type Environment struct {
	parent *Environment
	values map[VSym]interface{}
}

func (e *Environment) String() string {
	if e == nil {
		return "{}"
	}
	var buf bytes.Buffer
	var count int
	for count = 0; e != nil; count, e = count+1, e.parent {
		buf.WriteRune('{')
		for k, v := range e.values {
			buf.WriteString(string(k))
			buf.WriteRune(':')
			buf.WriteString(fmt.Sprintf("%s", v))
			buf.WriteString(", ")
		}
	}
	for ; count > 0; count-- {
		buf.WriteRune('}')
	}
	return buf.String()
}

func NewEnv(p *Environment, v map[VSym]interface{}) *Environment {
	return &Environment{parent: p, values: v}
}

func (env *Environment) Get(x VSym) interface{} {
loop:
	if val, ok := env.values[x]; ok {
		return val
	}
	if env.parent != nil {
		env = env.parent
		goto loop
	}
	panic("Unbound Symbol")
}

func (env *Environment) Set(x VSym, y interface{}) interface{} {
	env.values[x] = y
	return y
}

type Evaller func(interface{}, *Environment) interface{}

type Callable interface {
	Call(Evaller, *Environment, *VPair) (interface{}, *Environment, bool)
}

type NativeFn struct {
	Fn func(Evaller, *Environment, *VPair) (interface{}, *Environment, bool)
}

func (nfn NativeFn) Call(eval Evaller, env *Environment, x *VPair) (v interface{}, e *Environment, r bool) {
	v, e, r = nfn.Fn(eval, env, x)
	return
}
func (nfn NativeFn) String() string {
	return "<native code>"
}

func match_args(f *VPair, a *VPair) map[VSym]interface{} {
	m := make(map[VSym]interface{})
	for f != nil {
		if a == nil {
			panic("Too few arguments")
		}
		s, ok := f.Car.(VSym)
		if !ok {
			panic("Cannot bind to non-symbol")
		}
		m[s] = a.Car
		if fp, ok := f.Cdr.(*VPair); ok {
			f = fp
		}
		if ap, ok := a.Cdr.(*VPair); ok {
			a = ap
		}
	}
	return m
}

type Combiner struct {
	Stat_env *Environment
	Formals  *VPair
	Dyn_sym  VSym
	Body     interface{}
}

func (c *Combiner) Call(eval Evaller, dyn_env *Environment, args *VPair) (interface{}, *Environment, bool) {
	arg_map := match_args(c.Formals, args)
	arg_map[c.Dyn_sym] = dyn_env
	return c.Body, NewEnv(c.Stat_env, arg_map), true
}

func (c *Combiner) String() string {
	return "<combiner>"
}

type Applicative struct {
	Wrapper func(Evaller, *Environment, *VPair, *VPair) map[VSym]interface{}
	Vau     *Combiner
}

func (a *Applicative) Call(eval Evaller, dyn_env *Environment, args *VPair) (interface{}, *Environment, bool) {
	v := a.Vau
	return v.Body, NewEnv(v.Stat_env, a.Wrapper(eval, dyn_env, v.Formals, args)), true
}

func (a *Applicative) String() string {
	return "<applicative>"
}
