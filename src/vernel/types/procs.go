package types

import "fmt"

type Continuation struct {
	Fn func(*VPair) *Tail
}

func (k *Continuation) Call(_ Evaller, _ *Environment, _ *Continuation, x *VPair) *Tail {
	return k.Fn(x)
}

func (k *Continuation) String() string {
	return "<cont>"
}

type NativeFn struct {
	Fn func(Evaller, *Environment, *Continuation, *VPair) *Tail
}

func (nfn *NativeFn) Call(eval Evaller, env *Environment, k *Continuation, x *VPair) *Tail {
	return nfn.Fn(eval, env, k, x)
}

func (nfn *NativeFn) String() string {
	return "<native>"
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

func (c *Combiner) Call(_ Evaller, denv *Environment, k *Continuation, args *VPair) *Tail {
	arg_map := match_args(c.Formals, args)
	arg_map[c.Dsym] = WrapEnv(denv)
	return &Tail{c.Body, NewEnv(c.Cenv, arg_map), k}
}

func (c *Combiner) String() string {
	return "<combiner>"
}

type Applicative struct {
	Wrapper  func(Callable, Evaller, *Environment, *Continuation, *VPair) *Tail
	Internal Callable
}

func (a *Applicative) Call(eval Evaller, denv *Environment, k *Continuation, args *VPair) *Tail {
	return a.Wrapper(a.Internal, eval, denv, k, args)
}

func (a *Applicative) String() string {
	return fmt.Sprintf("<applicative: %s>", a.Internal)
}

var Top = &Continuation{
	func(args *VPair) *Tail {
		return &Tail{args.Car, nil, nil}
	},
}