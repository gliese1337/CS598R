package types

import "fmt"

type Continuation struct {
	Name string
	Fn   func(*Tail, *VPair) bool
}

func (k *Continuation) Call(_ Evaller, ctx *Tail, x *VPair) bool {
	return k.Fn(ctx, x)
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
	Fn   func(Evaller, *Tail, *VPair) bool
}

func (nfn *NativeFn) Call(eval Evaller, ctx *Tail, x *VPair) bool {
	return nfn.Fn(eval, ctx, x)
}
func (nfn *NativeFn) Strict(ctx *Tail) bool {
	ctx.Expr = nfn
	return false
}
func (nfn *NativeFn) String() string {
	return fmt.Sprintf("<native:%s>", nfn.Name)
}

func match_args(fs VValue, a *VPair) map[VSym]VValue {
	var m map[VSym]VValue
	switch f := fs.(type) {
	case *VPair:
		for f != nil {
			if a == nil {
				panic("Too few arguments")
			}
			s, ok := f.Car.(VSym)
			if !ok {
				panic("Cannot bind to non-symbol")
			}
			if string(s) != "##" {
				m[s] = a.Car
			}

			ap, ok := a.Cdr.(*VPair)
			switch fp := f.Cdr.(type) {
			case *VPair:
				f, a = fp, ap
			case VSym:
				if ok {
					if string(fp) != "##" {
						m[fp] = ap
					}
					f = nil
				}
			}
		}
	case VSym:
		if string(f) != "##" {
			m[f] = a
		}
	default:
		panic("Invalid formals!")
	}
	return m
}

type Combiner struct {
	Cenv    *Environment
	Formals VValue
	Dsym    VSym
	Body    *VPair
}

func (c *Combiner) Call(_ Evaller, ctx *Tail, args *VPair) bool {
	arg_map := match_args(c.Formals, args)
	if c.Body == nil {
		ctx.Expr = VNil
		return false
	}
	if string(c.Dsym) != "##" {
		arg_map[c.Dsym] = WrapEnv(ctx.Env)
	}
	senv, sk := NewEnv(c.Cenv, arg_map), ctx.K
	var eloop func(*Tail, *VPair) bool
	eloop = func(kctx *Tail, body *VPair) bool {
		var cfunc func(*Tail, *VPair) bool
		next_expr, ok := body.Cdr.(*VPair)
		if !ok {
			panic("Invalid Function Body")
		}
		if next_expr == nil {
			cfunc = func(nctx *Tail, va *VPair) bool {
				nctx.Expr, nctx.K = va.Car, sk
				return false
			}
		} else {
			cfunc = func(nctx *Tail, va *VPair) bool {
				return eloop(nctx, next_expr)
			}
		}
		kctx.Expr, kctx.Env, kctx.K = body.Car, senv, &Continuation{"arg", cfunc}
		return true
	}
	return eloop(ctx, c.Body)
}
func (c *Combiner) Strict(ctx *Tail) bool {
	ctx.Expr = c
	return false
}
func (c *Combiner) String() string {
	return "<combiner>"
}

type Applicative struct {
	Wrapper  func(Callable, Evaller, *Tail, *VPair) bool
	Internal Callable
}

func (a *Applicative) Call(eval Evaller, ctx *Tail, args *VPair) bool {
	return a.Wrapper(a.Internal, eval, ctx, args)
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
