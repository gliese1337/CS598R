package types

import (
	"bytes"
	"fmt"
)

type Environment struct {
	parent *Environment
	values map[VSym]VValue
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

func (env *Environment) Get(x VSym) VValue {
loop:
	if val, ok := env.values[x]; ok {
		return val
	}
	if env.parent != nil {
		env = env.parent
		goto loop
	}
	panic("Unbound Symbol: " + string(x))
}

func (env *Environment) Set(x VSym, y VValue) VValue {
	env.values[x] = y
	return y
}

func (env *Environment) Call(_ Evaller, ctx *Tail, args *VPair) bool {
	if args == nil {
		panic("No argument to evaluation")
	}
	ctx.Expr, ctx.Env = args.Car, env
	return true
}

func NewEnv(p *Environment, v map[VSym]VValue) *Environment {
	return &Environment{parent: p, values: v}
}

func WrapEnv(p *Environment) *Applicative {
	return &Applicative{func(_ Callable, _ Evaller, ctx *Tail, cargs *VPair) bool {
		if cargs == nil {
			ctx.Expr = VNil
		} else {
			sctx := *ctx
			ctx.Expr, ctx.K = cargs.Car, &Continuation{
				"eval",
				func(nctx *Tail, v *VPair) bool {
					evaluate := p.Call(nil, &sctx, v)
					*nctx = sctx
					return evaluate
				},
			}
		}
		return true
	}, p}
}
