package lib

import (
	. "vernel/types"
)

func vau(eval Evaller, clos_env *Environment, x *VPair) (interface{}, *Environment, bool) {
	if x == nil {
		panic("No Arguments to vau")
	}
	formals, ok := x.Car.(*VPair)
	if !ok {
		panic("Invalid Argument Declaration")
	}
	sym_rest, ok := x.Cdr.(*VPair)
	if !ok {
		panic("Invalid Arguments to vau")
	}
	dyn_sym, ok := sym_rest.Car.(VSym)
	if !ok {
		panic("Missing Dynamic Environment Binding")
	}
	rest, ok := sym_rest.Cdr.(*VPair)
	if !ok {
		panic("Missing Vau Expression Body")
	}
	return &Combiner{
		Stat_env: clos_env,
		Formals:  formals,
		Dyn_sym:  dyn_sym,
		Body:     rest.Car,
	}, nil, false
}

func rtlWrapper(eval Evaller, dyn_env *Environment, f *VPair, a *VPair) map[VSym]interface{} {
	m := make(map[VSym]interface{})
	for f != nil {
		if a == nil {
			panic("Too few arguments")
		}
		s, ok := f.Car.(VSym)
		if !ok {
			panic("Cannot bind to non-symbol")
		}
		m[s] = eval(a.Car, dyn_env)
		if fp, ok := f.Cdr.(*VPair); ok {
			f = fp
		}
		if ap, ok := a.Cdr.(*VPair); ok {
			a = ap
		}
	}
	return m
}

func wrap(eval Evaller, clos_env *Environment, x *VPair) (interface{}, *Environment, bool) {
	if x == nil {
		panic("No Arguments to vau")
	}
	formals, ok := x.Car.(*VPair)
	if !ok {
		panic("Invalid Argument Declaration")
	}
	sym_rest, ok := x.Cdr.(*VPair)
	if !ok {
		panic("Invalid Arguments to vau")
	}
	dyn_sym, ok := sym_rest.Car.(VSym)
	if !ok {
		panic("Missing Dynamic Environment Binding")
	}
	rest, ok := sym_rest.Cdr.(*VPair)
	if !ok {
		panic("Missing Vau Expression Body")
	}
	return &Combiner{
		Stat_env: clos_env,
		Formals:  formals,
		Dyn_sym:  dyn_sym,
		Body:     rest.Car,
	}, nil, false
}

func cons(eval Evaller, env *Environment, x *VPair) (interface{}, *Environment, bool) {
	if x == nil {
		panic("No Arguments to cons")
	}
	if cdr, ok := x.Cdr.(*VPair); ok {
		return &VPair{eval(x.Car, env), eval(cdr.Car, env)}, nil, false
	}
	panic("Invalid Arguments to cons")
}

func car(eval Evaller, env *Environment, x *VPair) (interface{}, *Environment, bool) {
	if x == nil {
		panic("No Argument to car")
	}
	if arg, ok := eval(x.Car, env).(*VPair); ok {
		if arg == nil {
			panic("Empty List Passed to car")
		}
		return arg.Car, nil, false
	}
	panic("Invalid Argument to car")
}

func cdr(eval Evaller, env *Environment, x *VPair) (interface{}, *Environment, bool) {
	if x == nil {
		panic("No Argument to cdr")
	}
	if arg, ok := eval(x.Car, env).(*VPair); ok {
		if arg == nil {
			panic("Empty List Passed to cdr")
		}
		return arg.Cdr, nil, false
	}
	panic("Invalid Argument to cdr")
}

var Standard = NewEnv(
	nil,
	map[VSym]interface{}{
		VSym("cons"): NativeFn{cons},
		VSym("car"):  NativeFn{car},
		VSym("cdr"):  NativeFn{cdr},
		VSym("vau"):  NativeFn{vau},
	})
