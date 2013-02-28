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
		Body:    rest,
	}
	return false
}

func rtlWrapper(internal Callable, eval Evaller, ctx *Tail, args *VPair) bool {
	//TODO: make thread safe
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
	/*copy cells up to the target, but keep the same tail; works as long as cells are immutable*/
	first := *arglist
	last := &first
	for i := 1; i <= idx; i++ {
		next := *last.Cdr.(*VPair) /*copy next cell*/
		last.Cdr = &next           /*hook up to previous copied cell*/
		last = &next
	}
	last.Car = val
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
	last_idx := count - 1
	blocked := make(map[*Tail]struct{})
	arglist := make([]VPair, count, count)
	c_lock := new(sync.RWMutex)

	make_reactivation := func(idx int) func(*Tail, *VPair) bool {
		reassign := func(nctx *Tail, vals *VPair) bool {
			*nctx = sctx
			nargs := copy_arglist(idx, vals.Car, &(arglist[0]))
			return internal.Call(eval, nctx, nargs)
		}
		return func(nctx *Tail, vals *VPair) bool {
			c_lock.RLock()
			if count > 0 {
				c_lock.RUnlock()
				context := Tail{nil, sctx.Env, &Continuation{"ArgK", reassign}}
				blocked[&context] = struct{}{}
				nctx.K = nil
				return false
			}
			c_lock.RUnlock()
			return reassign(nctx, vals)
		}
	}

	make_argk := func(idx int) *Continuation {
		var k Continuation
		var reactivate func(*Tail, *VPair) bool
		k_lock := new(sync.Mutex)
		activated := false
		k.Name = "ArgK"
		k.Fn = func(nctx *Tail, vals *VPair) bool {
			k_lock.Lock()
			if activated {
				k_lock.Unlock()
				return reactivate(nctx, vals)
			}
			activated = true
			k_lock.Unlock()
			arglist[idx].Car = vals.Car
			reactivate = make_reactivation(idx)
			k.Fn = reactivate
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
				go eval(context.Expr, context.Env, context.K)
			}
			*nctx = sctx
			return internal.Call(eval, nctx, &(arglist[0]))
		}
		return &k
	}

	//start new goroutine for all but one argument
	arglist[last_idx].Cdr = VNil
	for i := 0; i < last_idx; i++ {
		arglist[i].Cdr = &(arglist[i+1])
		switch a := args.Car.(type) {
		case *VPair:
			go eval(a, ctx.Env, make_argk(i))
		case VSym:
			go eval(a, ctx.Env, make_argk(i))
		default:
			arglist[i].Car = args.Car
			c_lock.Lock()
			count--
			c_lock.Unlock()
		}
		args = args.Cdr.(*VPair)
	}
	//reuse current goroutine for last argument
	ctx.Expr = args.Car
	ctx.K = make_argk(last_idx)
	return true
}

func basicWrapper(internal Callable, eval Evaller, ctx *Tail, args *VPair) bool {
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
	finished := false
	blocked := make(map[*Tail]struct{})
	arglist := make([]VPair, count, count)
	f_lock := new(sync.RWMutex)

	make_reactivation := func(idx int) func(*Tail, *VPair) bool {
		reassign := func(nctx *Tail, vals *VPair) bool {
			*nctx = sctx
			nargs := copy_arglist(idx, vals.Car, &(arglist[0]))
			return internal.Call(eval, nctx, nargs)
		}
		return func(nctx *Tail, vals *VPair) bool {
			f_lock.RLock()
			if finished {
				f_lock.RUnlock()
				return reassign(nctx, vals)
			}
			f_lock.RUnlock()
			context := Tail{nil, sctx.Env, &Continuation{"ArgK", reassign}}
			blocked[&context] = struct{}{}
			nctx.K = nil
			return false
		}
	}

	make_argk := func(idx int) *Continuation {
		var k Continuation
		var reactivate func(*Tail, *VPair) bool
		k_lock := new(sync.Mutex)
		activated := false
		k.Name = "ArgK"
		k.Fn = func(nctx *Tail, vals *VPair) bool {
			k_lock.Lock()
			if activated {
				k_lock.Unlock()
				return reactivate(nctx, vals)
			}
			activated = true
			k_lock.Unlock()
			arglist[idx].Car = vals.Car
			reactivate = make_reactivation(idx)
			k.Fn = reactivate
			nctx.K = nil
			return false
		}
		return &k
	}

	//randomize evaluation order
	//TODO: Somehow calculate optimal evaluation orders....
	var argset map[int]interface{}
	arglist[count-1].Cdr = VNil
	for i := 0; i < count; i++ {
		arglist[i].Cdr = &(arglist[i+1])
		argset[i] = args.Car
		args = args.Cdr.(*VPair)
	}
	for i, a := range argset {
		switch at := a.(type) {
		case *VPair:
			eval(at, ctx.Env, make_argk(i))
		case VSym:
			eval(at, ctx.Env, make_argk(i))
		default:
			arglist[i].Car = a
		}
	}
	f_lock.Lock()
	finished = true
	f_lock.Unlock()
	for context, _ := range blocked {
		go eval(context.Expr, context.Env, context.K)
	}
	return internal.Call(eval, ctx, &(arglist[0]))
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
				evaluate := qwrapf(nil, &sctx, v)
				*nctx = sctx
				return evaluate
			},
		}
		return true
	}, &NativeFn{"wrapper", qwrapf}}
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
