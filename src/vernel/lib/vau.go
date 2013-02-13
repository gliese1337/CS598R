package lib

import . "vernel/types"
import "sync"

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

func copy_arglist(idx int, val interface{}, arglist *VPair) *VPair {
	//copy cells up to the target, but keep the same tail; works as long as cells are immutable
	first := *arglist
	last := &first
	for i := 1; i <= idx; i++ {
		next := *last.Cdr
		last.Cdr = &next
		last = next
	}
	last.Car = vals.Car
	return &first
}

func syncWrapper(internal Callable, eval Evaller, ctx *Tail, args *VPair) bool {
	if args == nil {
		return internal.Call(eval, ctx, VNil)
	}
	count, ok := 0, true
	for a := args; a != nil && ok; a, ok = a.Cdr.(*VPair) {
		count++
	}
	if !ok {
		panic("Invalid Argument List")
	}
	
	sctx := *ctx
	last_idx := count-1
	blocked := make(map[*Tail]struct{})
	arglist := make([]VPair,count,count)
	c_lock := new(sync.RWMutex)
	
	reactivate := func(idx int) (func(*Tail, *VPair) bool) {
		reassign := func(nctx *Tail, vals *VPair) bool {
			*nctx = sctx
			nargs := copy_arglist(idx,vals.Car,&(arglist[0])
			return internal.Call(eval,nctx,nargs)
		}
		return func(nctx *Tail, vals *VPair) bool {
			c_lock.RLock()
			if count > 0 {
				c_lock.RUnlock()
				context := Tail{nil,sctx.Env,&Continuation{"ArgK",reassign}}
				blocked[&context] = struct{}
				nctx.K = nil
				return false
			}
			c_lock.RUnlock()
			return reassign(nctx,vals)
		}
	}
	
	make_argk := func(idx int) &Continuation {
		var k Continuation
		k.Name = "ArgK"
		k.Fn = func(nctx *Tail, vals *VPair) bool {
			arglist[idx].Car = vals.Car
			k.Fn = reactivate(idx)
			c_lock.Lock()
			count -= 1
			if count > 0 { //die silently
				c_lock.Unlock()
				nctx.K = nil
				return false
			}
			c_lock.Unlock()
			//last one done, evaluate the body and anybody who blocked
			for context, _ := range blocked {
				go eval(context.Expr,context.Env,context.K)
			}
			*nctx = sctx
			return internal.Call(eval,nctx,&(arglist[0]))
		}
		return &k
	}
	
	//start new goroutine for all but one argument
	arglist[last_idx].Cdr = VNil
	for i := 0; i < last_idx; i++ {
		arglist[i].Cdr = &(arglist[i+1])
		go eval(a.Car,ctx.Env,make_argk(i))
		args = args.Cdr.(*VPair)
	}
	//reuse current goroutine for last argument
	ctx.Expr = arglist[last_idx].Car
	ctx.K = make_argk(last_idx)	
	return true
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
