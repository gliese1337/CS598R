package lib

import . "vernel/types"

func vau(_ Evaller, cenv *Environment, k *Continuation, x *VPair) *Tail {
	if x == nil {
		panic("No Arguments to vau")
	}
	_, islist := x.Car.(*VPair)
	_, issym := x.Car.(VSym)
	if !islist || issym {
		panic("Invalid Argument Declaration")
	}
	sym_rest, ok := x.Cdr.(*VPair)
	if !ok {
		panic("Invalid Arguments to vau")
	}
	dsym, ok := sym_rest.Car.(VSym)
	if !ok {
		panic("Missing Dynamic Environment Binding")
	}
	rest, ok := sym_rest.Cdr.(*VPair)
	if rest == nil || !ok {
		panic("Missing Vau Expression Body")
	}
	return k.Fn(&VPair{&Combiner{
		Cenv:    cenv,
		Formals: x.Car,
		Dsym:    dsym,
		Body:    rest.Car,
	}, VNil})
}

func rtlWrapper(internal Callable, eval Evaller, env *Environment, k *Continuation, args *VPair) *Tail {
	/*map arglist with eval function*/
	if args == nil {
		return internal.Call(eval, env, k, VNil)
	}

	argv := &VPair{nil, VNil}

	var map_loop func(*VPair, *VPair) *Tail
	map_loop = func(oa *VPair, na *VPair) *Tail {
		ccall := false
		return &Tail{oa.Car, env, &Continuation{func(va *VPair) *Tail {
			if !ccall {
				ccall = true
				na.Car = va.Car
				if next_arg, ok := oa.Cdr.(*VPair); ok {
					if next_arg == nil {
						na.Cdr = VNil
						return internal.Call(eval, env, k, argv)
					}
					next_slot := &VPair{nil, nil}
					na.Cdr = next_slot
					return map_loop(next_arg, next_slot)
				} else {
					panic("Invalid Argument List")
				}
			}
			panic("Continuation Activation Not Implemented")
			nargv := argv
			return internal.Call(eval, env, k, nargv)
		}}}
	}
	return map_loop(args, argv)
}

func wrap_gen(fn func(Callable, Evaller, *Environment, *Continuation, *VPair) *Tail) Callable {
	return &NativeFn{func(eval Evaller, env *Environment, k *Continuation, x *VPair) *Tail {
		if x == nil {
			panic("No Argument to wrap")
		}
		return &Tail{x.Car, env, &Continuation{func(v *VPair) *Tail {
			if proc, ok := v.Car.(Callable); ok {
				return k.Fn(&VPair{&Applicative{fn, proc}, VNil})
			}
			panic("Non-callable passed to wrap")
		}}}
	}}
}

func unwrap(eval Evaller, env *Environment, k *Continuation, x *VPair) *Tail {
	if x == nil {
		panic("No Arguments unwrap")
	}
	return &Tail{x.Car, env, &Continuation{func(v *VPair) *Tail {
		if arg, ok := v.Car.(*Applicative); ok {
			return k.Fn(&VPair{arg.Internal, VNil})
		}
		panic("Can't unwrap non-applicative")
	}}}
}
