package lib

import . "vernel/types"

func vau(_ Evaller, ctx *Tail, x *VPair) bool {
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
	return false
}

func rtlWrapper(internal Callable, eval Evaller, ctx *Tail, args *VPair) bool {
	/*map arglist with eval function*/
	if args == nil {
		return internal.Call(eval, ctx, VNil)
	}
	senv, sk := ctx.Env, ctx.K
	argv := VPair{nil, VNil}
	var mloop func(*Tail, *VPair, *VPair) bool
	mloop = func(kctx *Tail, oa *VPair, na *VPair) bool {
		kctx.Expr = oa.Car
		kctx.Env = senv //TODO: Why is this necessary?
		kctx.K = &Continuation{"arg", func(nctx *Tail, va *VPair) bool {
			na.Car = va.Car
			if next_arg, ok := oa.Cdr.(*VPair); ok {
				if next_arg == nil {
					nctx.Env, nctx.K = senv, sk
					return internal.Call(eval, nctx, &argv)
				} else {
					next_slot := &VPair{nil, VNil}
					na.Cdr = next_slot
					return mloop(nctx, next_arg, next_slot)
				}
			}
			panic("Invalid Argument List")
		}}
		return true
	}
	return mloop(ctx, args, &argv)
}

func wrap_gen(fn func(Callable, Evaller, *Tail, *VPair) bool) Callable {
	qwrapf := func(_ Evaller, ctx *Tail, x *VPair) bool {
		if x == nil {
			panic("No Argument to wrap")
		}
		if proc, ok := x.Car.(Callable); ok {
			ctx.Expr = &Applicative{fn, proc}
			return false
		}
		panic("Non-callable passed to wrap")
	}
	return &Applicative{func(_ Callable, _ Evaller, ctx *Tail, cargs *VPair) bool {
		if cargs == nil {
			ctx.Expr = VNil
		} else {
			ctx.Expr = cargs.Car
		}
		sctx := *ctx
		ctx.K = &Continuation{
			"wrap",
			func(nctx *Tail, v *VPair) bool {
				evaluate := qwrapf(nil,&sctx,v)
				*nctx = sctx
				return evaluate
			},
		}
		return true
	},&NativeFn{"wrapper",qwrapf}}
}

func qunwrap(eval Evaller, ctx *Tail, x *VPair) bool {
	if x == nil {
		panic("No Arguments unwrap")
	}
	if arg, ok := x.Car.(*Applicative); ok {
		ctx.Expr = arg.Internal
		return false
	}
	panic("Can't unwrap non-applicative")
}
