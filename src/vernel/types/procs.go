package types

type NativeFn struct {
	Fn func(Evaller, *Environment, *VPair) (interface{}, *Environment, bool)
}

func (nfn NativeFn) Call(eval Evaller, env *Environment, x *VPair) (b interface{}, e *Environment, r bool) {
	b, e, r = nfn.Fn(eval, env, x)
	return
}

func (nfn NativeFn) String() string {
	return "<native code>"
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
	Stat_env *Environment
	Formals  *VPair
	Dyn_sym  VSym
	Body     interface{}
}

func (c *Combiner) Call(eval Evaller, dyn_env *Environment, args *VPair) (interface{}, *Environment, bool) {
	arg_map := match_args(c.Formals, args)
	arg_map[c.Dyn_sym] = dyn_env
	return c.Body, NewEnv(c.Stat_env, arg_map), true
}

func (c *Combiner) String() string {
	return "<combiner>"
}

type Applicative struct {
	Wrapper  func(Callable, Evaller, *Environment, *VPair) (interface{}, *Environment, bool)
	Internal Callable
}

func (a *Applicative) Call(eval Evaller, denv *Environment, args *VPair) (b interface{}, e *Environment, r bool) {
	b, e, r = a.Wrapper(a.Internal, eval, denv, args)
	return
}

func (a *Applicative) String() string {
	return "<applicative>"
}
