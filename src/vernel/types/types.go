package types

type VSym string
type VBool bool

type VPair struct {
	Car interface{}
	Cdr interface{}
}

var VNil *VPair = nil

type Environment struct {
	parent *Environment
	values map[VSym] interface{}
}

func NewEnv(p *Environment, v map[VSym] interface{}) *Environment {
	return &Environment{parent:p,values:v}
}

func (env *Environment) Get(x VSym) interface{} {
	val, ok := env.values[x]
	if ok { return val }
	if env.parent != nil { return env.parent.Get(x) }
	panic("Unbound Symbol")
}

func (env *Environment) Set(x VSym, y interface{}) interface{} {
	env.values[x] = y
	return y
}

type Tail struct {
	Expr interface{}
	Env *Environment
}

type Evaller func(interface{},*Environment) interface{}

type Callable interface {
	Call(Evaller,*Environment,*VPair) interface{}
}

type NativeFn struct {
	Fn func (Evaller,*Environment,*VPair) interface{}
}
func (nfn NativeFn) Call(eval Evaller, env *Environment, x *VPair) interface{} {
	return nfn.Fn(eval, env, x)
}

func match_args(f *VPair, a *VPair) map[VSym]interface{} {
	m := make(map[VSym]interface{})
	for f != nil {
		if a == nil { panic("Too few arguments") }
		s, ok := f.Car.(VSym)
		if !ok { panic("Cannot bind to non-symbol") }
		m[s] = a.Car
		if fp, ok := f.Cdr.(*VPair); ok { f = fp }
		if ap, ok := a.Cdr.(*VPair); ok { a = ap }
	}
	return m
}

type Combiner struct {
	Stat_env *Environment
	Formals *VPair
	Dyn_sym VSym
	Body interface{}
}

func (c *Combiner) Call(eval Evaller, dyn_env *Environment, args *VPair) interface{} {
	arg_map := match_args(c.Formals, args)
	arg_map[c.Dyn_sym] = dyn_env
	return Tail{c.Body, NewEnv(c.Stat_env, arg_map)}
}

type Applicative struct {
	Wrapper func(Evaller, *Environment, *VPair, *VPair) map[VSym]interface{}
	Vau *Combiner
}

func (a *Applicative) Call(eval Evaller, dyn_env *Environment, args *VPair) interface{} {
	v := a.Vau
	return Tail{v.Body, NewEnv(v.Stat_env, a.Wrapper(eval, dyn_env, v.Formals, args))}
}
