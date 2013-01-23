package types

import (
	"bytes"
	"fmt"
)

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

func (env *Environment) Call(eval Evaller, dyn_env *Environment, args *VPair) (interface{}, *Environment, bool) {
	if args == nil {
		return VNil, nil, false
	}
	return args.Car, env, true
}

func NewEnv(p *Environment, v map[VSym]interface{}) *Environment {
	return &Environment{parent: p, values: v}
}
