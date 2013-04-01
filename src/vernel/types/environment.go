package types

import (
	"bytes"
	"fmt"
)

type Environment struct {
	parent *Environment
	values map[VSym]VValue
}

func (e *Environment) GetSize(seen map[VValue]struct{}) int {
	if _, ok := seen[e]; e == nil || ok {
		return 0
	}
	total := e.parent.GetSize(seen)
	for _, v := range e.values {
		total += 1 + v.GetSize(seen)
	}
	return total
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
	for env != nil {
		if val, ok := env.values[x]; ok {
			return val
		}
		env = env.parent
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
			sexpr, senv, sk := ctx.Expr, ctx.Env, ctx.K
			ctx.Expr, ctx.K = cargs.Car, &Continuation{
				"eval",
				func(nctx *Tail, v *VPair) bool {
					sctx := Tail{sexpr, senv, sk, nctx.Time}
					evaluate := p.Call(nil, &sctx, v)
					*nctx = sctx
					return evaluate
				},
				[]VValue{p, sexpr, senv, sk},
			}
		}
		return true
	}, p}
}
