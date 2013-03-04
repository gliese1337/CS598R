package types

import (
	"bytes"
	"fmt"
)

type Environment struct {
	parent *Environment
	values map[VSym]VValue
}

func (e *Environment) Strict(ctx *Tail) bool {
	ctx.Expr = e
	return false
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
	if string(x) == "##" {
		panic("## always unbound")
	}
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
	if string(x) != "##" {
		env.values[x] = y
	}
	return y
}

func (env *Environment) Arity() (int,bool) {
	return 1,false
}
func (env *Environment) Call(_ Evaller, ctx *Tail, args ...VValue) bool {
	ctx.Expr, ctx.Env = args[0], env
	return true
}

func NewEnv(p *Environment, v map[VSym]VValue) *Environment {
	return &Environment{parent: p, values: v}
}

func WrapEnv(p *Environment) *Applicative {
	return &Applicative{
		Internal: p,
		Wrapper: func(env Callable, _ Evaller, ctx *Tail, cargs ...VValue) bool {
			sctx := *ctx
			ctx.Expr, ctx.K = cargs[0], &Continuation{
				Name: "eval",
				Argc: 1,
				Variadic: false,
				Fn: func(nctx *Tail, v ...VValue) bool {
					evaluate := env.Call(nil, &sctx, v...)
					*nctx = sctx
					return evaluate
				},
			}
			return true
		},
	}
}
