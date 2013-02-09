package lib

import . "vernel/types"

func vau(_ Evaller, ctx *Tail, x *VPair) {
	if x == nil {
		panic("No Arguments to vau")
	}
	_, islist := x.Car.(*VPair)
	_, issym := x.Car.(VSym)
	if !(islist || issym) {
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
	ctx.Expr = &Combiner{
		Cenv:    ctx.Env,
		Formals: x.Car,
		Dsym:    dsym,
		Body:    rest.Car,
	}
}

func rtlWrapper(internal Callable, eval Evaller, ctx *Tail, args *VPair) {
	/*map arglist with eval function*/
	if args == nil {
		internal.Call(eval, ctx, VNil)
		return
	}
	senv, sk := ctx.Env, ctx.K
	argv := VPair{nil, VNil}
	var mloop func(*Tail, *VPair, *VPair)
	mloop = func(kctx *Tail, oa *VPair, na *VPair) {
		kctx.Expr = oa.Car
		kctx.Env = senv
		kctx.K = &Continuation{"arg", func(nctx *Tail, va *VPair) {
			na.Car = va.Car
			if next_arg, ok := oa.Cdr.(*VPair); ok {
				if next_arg == nil {
					nctx.Env, nctx.K = senv, sk
					internal.Call(eval, nctx, &argv)
				} else {
					next_slot := &VPair{nil, VNil}
					na.Cdr = next_slot
					mloop(nctx, next_arg, next_slot)
				}
			} else {
				panic("Invalid Argument List")
			}
		}}
	}
	mloop(ctx, args, &argv)
}

func wrap_gen(fn func(Callable, Evaller, *Tail, *VPair)) Callable {
	return &NativeFn{"wrapper", func(eval Evaller, ctx *Tail, x *VPair) {
		if x == nil {
			panic("No Argument to wrap")
		}
		sk := ctx.K
		ctx.Expr, ctx.K = x.Car, &Continuation{"wrapped", func(nctx *Tail, v *VPair) {
			if proc, ok := v.Car.(Callable); ok {
				sk.Fn(nctx, &VPair{&Applicative{fn, proc}, VNil})
			} else {
				panic("Non-callable passed to wrap")
			}
		}}
	}}
}

func qunwrap(eval Evaller, ctx *Tail, x *VPair) {
	if x == nil {
		panic("No Arguments unwrap")
	}
	if arg, ok := x.Car.(*Applicative); ok {
		ctx.Return(&VPair{arg.Internal, VNil})
	} else {
		panic("Can't unwrap non-applicative")
	}
}
