package types

import "fmt"

type Continuation struct {
	Name string
	Fn   func(*Tail, *VPair) bool
}

func (k *Continuation) Call(_ Evaller, ctx *Tail, x *VPair) bool {
	return k.Fn(ctx, x)
}

func (k *Continuation) String() string {
	return fmt.Sprintf("<cont:%s>", k.Name)
}

var Top = &Continuation{
	"Top",
	func(ctx *Tail, args *VPair) bool {
		ctx.Expr, ctx.K = args.Car, nil
		return false
	},
}

type NativeFn struct {
	Name string
	Fn   func(Evaller, *Tail, *VPair) bool
}

func (nfn *NativeFn) Call(eval Evaller, ctx *Tail, x *VPair) bool {
	return nfn.Fn(eval, ctx, x)
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
	Formals interface{}
	Dsym    VSym
	Body    *VPair
}

func (c *Combiner) Call(_ Evaller, ctx *Tail, args *VPair) bool {
	arg_map := match_args(c.Formals, args)
	if string(c.Dsym) != "##" {
		arg_map[c.Dsym] = WrapEnv(ctx.Env)
	}
	E := NewEnv(c.Cenv, arg_map)
	B := c.Body
	//TODO: Auto-sequence
	ctx.Expr, ctx.Env = B.Car, E
	return true
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

func (a *Applicative) String() string {
	if _, ok := a.Internal.(*Environment); ok {
		return "<applicative: Env>"
	}
	return fmt.Sprintf("<applicative: %s>", a.Internal)
}
