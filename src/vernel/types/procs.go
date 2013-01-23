package types

type NativeFn struct {
	Fn func(Evaller, *Environment, *VPair) (interface{}, *Environment, bool)
}

func (nfn NativeFn) Call(eval Evaller, env *Environment, x *VPair) (v interface{}, e *Environment, r bool) {
	v, e, r = nfn.Fn(eval, env, x)
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
	Wrapper func(Evaller, *Environment, *VPair, *VPair) map[VSym]interface{}
	Vau     *Combiner
}

func (a *Applicative) Call(eval Evaller, dyn_env *Environment, args *VPair) (interface{}, *Environment, bool) {
	v := a.Vau
	return v.Body, NewEnv(v.Stat_env, a.Wrapper(eval, dyn_env, v.Formals, args)), true
}

func (a *Applicative) String() string {
	return "<applicative>"
}
