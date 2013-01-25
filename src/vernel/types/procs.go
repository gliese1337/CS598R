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
	Cenv    *Environment
	Formals *VPair
	Dsym    VSym
	Body    interface{}
}

func (c *Combiner) Call(_ Evaller, denv *Environment, k *Continuation, args *VPair) *Tail {
	arg_map := match_args(c.Formals, args)
	arg_map[c.Dsym] = &Applicative{
		func(_ Callable, _ Evaller, ce *Environment, ck *Continuation, cargs *VPair) *Tail {
			fmt.Printf("Env args: %s\n", cargs)
			if cargs == nil {
				return &Tail{VNil, ce, ck}
			}
			return &Tail{cargs.Car, ce, &Continuation{
				func(v *VPair) *Tail { return denv.Call(nil, ce, ck, v) },
			}}
		},
		denv,
	}
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
