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

func rtlWrapper(internal Callable, eval Evaller, dyn_env *Environment, args *VPair) (b interface{}, e *Environment, r bool) {
	/*map arglist with eval function*/
	var new_args *VPair
	b, e, r = internal.Call(eval, dyn_env, new_args)
	return
}

func wrap(eval Evaller, env *Environment, x *VPair) (interface{}, *Environment, bool) {
	if x == nil {
		panic("No Argument to wrap")
	}
	proc, ok := eval(x.Car, env).(Callable)
	if !ok {
		panic("Non-proc passed to wrap")
	}
	return &Applicative{
		Wrapper:  rtlWrapper,
		Internal: proc,
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

func def(eval Evaller, env *Environment, x *VPair) (interface{}, *Environment, bool) {
	if x == nil {
		panic("No Arguments to def")
	}
	sym, ok := x.Car.(VSym)
	if !ok {
		panic("Cannot define non-symbol")
	}
	rest, ok := x.Cdr.(*VPair)
	if !ok {
		panic("Non-list argument to def")
	}
	var val interface{}
	if rest == nil {
		val = VNil
	} else {
		val = eval(rest.Car, env)
	}
	env.Set(sym, val)
	return val, nil, false
}

var Standard = NewEnv(
	nil,
	map[VSym]interface{}{
		VSym("def"):  NativeFn{def},
		VSym("cons"): NativeFn{cons},
		VSym("car"):  NativeFn{car},
		VSym("cdr"):  NativeFn{cdr},
		VSym("vau"):  NativeFn{vau},
		VSym("wrap"): NativeFn{wrap},
	})
