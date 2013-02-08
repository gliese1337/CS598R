package types

import "fmt"

type Continuation struct {
	Name string
	Fn   func(*Tail, *VPair)
}

func (k *Continuation) Call(_ Evaller, ctx *Tail, x *VPair) {
	k.Fn(ctx, x)
}

func (k *Continuation) String() string {
	return fmt.Sprintf("<cont:%s>", k.Name)
}

var Top = &Continuation{
	"Top",
	func(ctx *Tail, args *VPair) {
		ctx.Expr, ctx.K = args.Car, nil
	},
}

type NativeFn struct {
	Name string
	Fn   func(Evaller, *Tail, *VPair)
}

func (nfn *NativeFn) Call(eval Evaller, ctx *Tail, x *VPair) {
	nfn.Fn(eval, ctx, x)
}

func (nfn *NativeFn) String() string {
	return fmt.Sprintf("<native:%s>", nfn.Name)
}

func match_args(fs interface{}, a *VPair) map[VSym]interface{} {
	m := make(map[VSym]interface{})
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
			m[s] = a.Car

			ap, ok := a.Cdr.(*VPair)
			switch fp := f.Cdr.(type) {
			case *VPair:
				f, a = fp, ap
			case VSym:
				if ok {
					m[fp] = ap
					f = nil
				}
			}
		}
	case VSym:
		m[f] = a
	default:
		panic("Invalid formals!")
	}
	return m
}

type Combiner struct {
	Cenv    *Environment
	Formals interface{}
	Dsym    VSym
	Body    interface{}
}

func (c *Combiner) Call(_ Evaller, ctx *Tail, args *VPair) {
	arg_map := match_args(c.Formals, args)
	arg_map[c.Dsym] = WrapEnv(ctx.Env)
	ctx.Expr, ctx.Env = c.Body, NewEnv(c.Cenv, arg_map)
}

func (c *Combiner) String() string {
	return "<combiner>"
}

type Applicative struct {
	Wrapper  func(Callable, Evaller, *Tail, *VPair)
	Internal Callable
}

func (a *Applicative) Call(eval Evaller, ctx *Tail, args *VPair) {
	a.Wrapper(a.Internal, eval, ctx, args)
}

func (a *Applicative) String() string {
	if _, ok := a.Internal.(*Environment); ok {
		return "<applicative: Env>"
	}
	return fmt.Sprintf("<applicative: %s>", a.Internal)
}
