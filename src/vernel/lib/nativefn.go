package lib

import . "vernel/types"
import "fmt"

func bindcc(_ Evaller, senv *Environment, k *Continuation, x *VPair) *Tail {
	if x == nil {
		panic("No arguments to bind/cc")
	}
	k_sym, ok := x.Car.(VSym)
	if !ok {
		panic("Cannot bind to non-symbol")
	}
	body, ok := x.Cdr.(*VPair)
	if !ok {
		panic("No body provided to bind/cc")
	}
	return &Tail{body.Car, NewEnv(senv, map[VSym]interface{}{
		k_sym: &Applicative{func(_ Callable, _ Evaller, cenv *Environment, _ *Continuation, args *VPair) *Tail {
			if args == nil {
				return k.Fn(VNil)
			}
			return &Tail{args.Car, cenv, k}
		}, k},
	}), k}
}

func qcons(_ Evaller, _ *Environment, k *Continuation, x *VPair) *Tail {
	if x == nil {
		panic("No Arguments to cons")
	}
	if cdr, ok := x.Cdr.(*VPair); ok {
		if cdr == nil {
			panic("Too few arguments to cons")
		}
		return k.Fn(&VPair{&VPair{x.Car, cdr.Car}, VNil})
	}
	panic("Invalid Arguments to cons")
}

func qcar(_ Evaller, _ *Environment, k *Continuation, x *VPair) *Tail {
	if x == nil {
		panic("No Argument to car")
	}
	if arg, ok := x.Car.(*VPair); ok {
		if arg == nil {
			panic("Empty List Passed to car")
		}
		return k.Fn(&VPair{arg.Car, VNil})
	}
	panic("Invalid Argument to car")
}

func qcdr(_ Evaller, _ *Environment, k *Continuation, x *VPair) *Tail {
	if x == nil {
		panic("No Argument to cdr")
	}
	if arg, ok := x.Car.(*VPair); ok {
		if arg == nil {
			panic("Empty List Passed to cdr")
		}
		return k.Fn(&VPair{arg.Cdr, VNil})
	}
	panic("Invalid Argument to cdr")
}

func last(_ Evaller, _ *Environment, k *Continuation, x *VPair) *Tail {
	var nx *VPair
	var ok bool
	for ; x != nil; x = nx {
		if nx, ok = x.Cdr.(*VPair); ok {
			if nx == nil {
				return k.Fn(&VPair{x, VNil})
			}
		} else {
			panic("Invalid Argument List")
		}
	}
	return k.Fn(VNil)
}

func qlist(_ Evaller, _ *Environment, k *Continuation, x *VPair) *Tail {
	return k.Fn(&VPair{x, VNil})
}

func def(_ Evaller, env *Environment, k *Continuation, x *VPair) *Tail {
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
		val = rest.Car
	}
	return &Tail{val, env, &Continuation{
		func(args *VPair) *Tail {
			env.Set(sym, args.Car)
			return k.Fn(&VPair{args.Car, VNil})
		},
	}}
}

func qprint(_ Evaller, env *Environment, k *Continuation, x *VPair) *Tail {
	var val interface{}
	if x == nil {
		val = VNil
	} else {
		val = x.Car
	}
	fmt.Printf("%s", val)
	return &Tail{val, env, k}
}
