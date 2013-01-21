package lib

import (
        . "vernel/types"
)

func vau(eval Evaller, clos_env *Environment, x *VPair) interface{} {
	formals, ok := x.Car.(*VPair)
	if !ok { panic("Invalid Argument Declaration") }
	sym_body, ok := x.Cdr.(*VPair)
	if !ok { panic("Invalid Arguments to Vau") }
	dyn_sym, ok := sym_body.Car.(VSym)
	if !ok { panic("Missing Dynamic Environment Binding") }
	return &Combiner{
		Stat_env: clos_env,
		Formals: formals,
		Dyn_sym: dyn_sym,
		Body: sym_body.Cdr,
	}
}

func cons(eval Evaller, env *Environment, x *VPair) interface{}{
	if cdr, ok := x.Cdr.(*VPair); ok {
		return VPair{eval(x.Car,env),eval(cdr.Car,env)}
	}
	panic("Invalid Arguments to cons")
	return nil
}

func car(eval Evaller, env *Environment, x *VPair) interface{}{
	if arg, ok := eval(x.Car,env).(*VPair); ok {
		return arg.Car;
	}
	panic("Invalid Arguments to car")
	return nil
}

func cdr(eval Evaller, env *Environment, x *VPair) interface{}{
	if arg, ok := eval(x.Car,env).(*VPair); ok {
		return arg.Cdr;
	}
	panic("Invalid Arguments to cdr")
	return nil
}

var Standard = NewEnv(
	nil,
	map[VSym] interface{} {
	VSym("cons"): NativeFn{cons},
	VSym("car"): NativeFn{car},
	VSym("cdr"): NativeFn{cdr},
	VSym("vau"): NativeFn{vau},
})
