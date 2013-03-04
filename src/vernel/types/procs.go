package types

import "fmt"

type Continuation struct {
	Name string
	Argc int
	Variadic bool
	Fn   func(*Tail, ...VValue) bool
}
func (k *Continuation) Arity() (int,bool){
	return k.Argc,k.Variadic
}
func (k *Continuation) Call(_ Evaller, ctx *Tail, x ...VValue) bool {
	if len(x) < k.Argc || (!k.Variadic && len(x) > k.Argc) {
		panic(fmt.Sprintf("%d values in %d-value context",len(x),k.Argc))
	}
	return k.Fn(ctx, x...)
}
func (k *Continuation) Strict(ctx *Tail) bool {
	ctx.Expr = k
	return false
}
func (k *Continuation) String() string {
	return fmt.Sprintf("<cont:%s>", k.Name)
}

type NativeFn struct {
	Name string
	Argc int
	Variadic bool
	Fn   func(Evaller, *Tail, ...VValue) bool
}
func (nfn *NativeFn) Arity() (int,bool) {
	return nfn.Argc,nfn.Variadic
}
func (nfn *NativeFn) Call(eval Evaller, ctx *Tail, x ...VValue) bool {
	return nfn.Fn(eval, ctx, x...)
}
func (nfn *NativeFn) Strict(ctx *Tail) bool {
	ctx.Expr = nfn
	return false
}
func (nfn *NativeFn) String() string {
	return fmt.Sprintf("<native:%s>", nfn.Name)
}

func match_args(fs []VValue, a []VValue) map[VSym]VValue {
	var m map[VSym]VValue
	for i, s := range fs {
		switch st := s.(type) {
		case VSym:
			if string(st) != "##" {
				m[st] = a[i]
			}
		case *VPair:
			panic("Destructuring not yet implemented")
		default:
			panic("Invalid binding")
		}
	}
	return m
}

type Combiner struct {
	Cenv    *Environment
	Formals []VValue
	Dsym    VSym
	Variadic bool
	Body    []VValue
}
func (c *Combiner) Arity() (int,bool) {
	if c.Variadic {
		return len(c.Formals)-1,true
	}
	return len(c.Formals),false
}
func (c *Combiner) Call(_ Evaller, ctx *Tail, args ...VValue) bool {
	arg_map := match_args(c.Formals, args)
	if len(c.Body) == 0 {
		ctx.Expr = VNil
		return false
	}
	if string(c.Dsym) != "##" {
		arg_map[c.Dsym] = WrapEnv(ctx.Env)
	}
	senv, sk := NewEnv(c.Cenv, arg_map), ctx.K
	var eloop func(*Tail, ...VValue) bool
	eloop = func(kctx *Tail, body ...VValue) bool {
		var cfunc func(*Tail, ...VValue) bool
		if len(body) == 1 {
			cfunc = func(nctx *Tail, va ...VValue) bool {
				nctx.Expr, nctx.K = va[0], sk
				return false
			}
		} else {
			cfunc = func(nctx *Tail, _ ...VValue) bool {
				return eloop(nctx, body[1:]...)
			}
		}
		kctx.Expr, kctx.Env = body[0], senv
		kctx.K = &Continuation{
			Name: "arg",
			Argc: 1,
			Variadic: false,
			Fn: cfunc,
		}
		return true
	}
	return eloop(ctx, c.Body...)
}
func (c *Combiner) Strict(ctx *Tail) bool {
	ctx.Expr = c
	return false
}
func (c *Combiner) String() string {
	return "<combiner>"
}

type Applicative struct {
	Wrapper  func(Callable, Evaller, *Tail, ...VValue) bool
	Internal Callable
}
func (a *Applicative) Arity() (int,bool) {
	return a.Internal.Arity()
}
func (a *Applicative) Call(eval Evaller, ctx *Tail, args ...VValue) bool {
	return a.Wrapper(a.Internal, eval, ctx, args...)
}
func (a *Applicative) Strict(ctx *Tail) bool {
	ctx.Expr = a
	return false
}
func (a *Applicative) String() string {
	if _, ok := a.Internal.(*Environment); ok {
		return "<applicative: Env>"
	}
	return fmt.Sprintf("<applicative: %s>", a.Internal)
}
