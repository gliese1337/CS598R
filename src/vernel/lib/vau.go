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
	if !ok {
		panic("Invalid Vau Expression Body")
	}
	ctx.Expr = &Combiner{
		Cenv:    ctx.Env,
		Formals: x.Car,
		Dsym:    dsym,
		Body:    rest,
	}
	return false
}

func copy_arglist(idx int, arglist *VPair) (*VPair, *VPair) {
	/*copy cells up to the target, but keep the same tail; works as long as cells are immutable*/
	first := *arglist
	last := &first
	for i := 1; i <= idx; i++ {
		next := *last.Cdr.(*VPair) /*copy next cell*/
		last.Cdr = &next           /*hook up to previous copied cell*/
		last = &next
	}
	return &first, last
}

func ltrWrapper(internal Callable, eval Evaller, ctx *Tail, args *VPair) bool {
	if args == nil {
		return internal.Call(eval, ctx, VNil)
	}
	senv, sk := ctx.Env, ctx.K
	var argloop func(*Tail, *VPair, int, *VPair, *VPair) bool
	argloop = func(kctx *Tail, argv *VPair, depth int, oa *VPair, na *VPair) bool {
		var next_call func(*Tail, *VPair, *VPair) bool
		next_arg, ok := oa.Cdr.(*VPair)
		if !ok {
			panic("Invalid Argument List")
		}
		c_lock := new(sync.Mutex)
		called := false
		if next_arg == nil {
			next_call = func(nctx *Tail, nargv *VPair, _ *VPair) bool {
				nctx.Env, nctx.K = senv, sk
				return internal.Call(eval, nctx, nargv)
			}
		} else {
			next_call = func(nctx *Tail, nargv *VPair, slot *VPair) bool {
				next_slot := &VPair{nil, VNil}
				slot.Cdr = next_slot
				return argloop(nctx, nargv, depth+1, next_arg, next_slot)
			}
		}

		//TODO: Why is restoring the environment necessary?
		kctx.Expr, kctx.Env = oa.Car, senv
		kctx.K = &Continuation{"arg", func(nctx *Tail, va *VPair) bool {
			var slot, nargv *VPair
			c_lock.Lock()
			if called {
				c_lock.Unlock()
				nargv, slot = copy_arglist(depth, argv)
			} else {
				called = true
				c_lock.Unlock()
				nargv, slot = argv, na
			}
			slot.Car = va.Car
			return next_call(nctx, nargv, slot)
		}}
		return true
	}
	nargs := VPair{nil, VNil}
	return argloop(ctx, &nargs, 0, args, &nargs)
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
			nargs, slot := copy_arglist(idx, &(arglist[0]))
			slot.Car = vals.Car
			return internal.Call(eval, nctx, nargs)
		}
		return func(nctx *Tail, vals *VPair) bool {
			c_lock.RLock()
			if count > 0 {
				c_lock.RUnlock()
				blocked[&Tail{vals.Car, sctx.Env, &Continuation{"ArgK", reassign}}] = struct{}{}
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
				delete(blocked, context)
				go eval(context, false)
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
		c_lock.Lock()
		switch a := args.Car.(type) {
		case *VPair:
			go eval(&Tail{a, ctx.Env, make_argk(i)}, true)
		case VSym:
			arglist[i].Car = ctx.Env.Get(a)
			count--
		default:
			arglist[i].Car = args.Car
			count--
		}
		c_lock.Unlock()
		args = args.Cdr.(*VPair)
	}
	//reuse current goroutine for last argument
	ctx.Expr = args.Car
	ctx.K = make_argk(last_idx)
	return true
}

func futureWrapper(internal Callable, eval Evaller, ctx *Tail, args *VPair) bool {
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
	blocked := make(map[*Tail]struct{})
	arglist := make([]VPair, count, count)
	done := false
	d_lock := new(sync.RWMutex)

	make_reactivation := func(idx int) func(*Tail, *VPair) bool {
		reassign := func(nctx *Tail, vals *VPair) bool {
			*nctx = sctx
			nargs, slot := copy_arglist(idx, &(arglist[0]))
			slot.Car = vals.Car
			return internal.Call(eval, nctx, nargs)
		}
		return func(nctx *Tail, vals *VPair) bool {
			d_lock.RLock()
			if done {
				d_lock.RUnlock()
				return reassign(nctx, vals)
			}
			d_lock.RUnlock()
			blocked[&Tail{vals.Car, sctx.Env, &Continuation{"ArgK", reassign}}] = struct{}{}
			nctx.K = nil
			return false
		}
	}

	make_argk := func(idx int, f *Future) *Continuation {
		var k Continuation
		var k_lock sync.Mutex
		var reactivate func(*Tail, *VPair) bool
		k.Name = "ArgK"
		k.Fn = func(nctx *Tail, vals *VPair) bool {
			k_lock.Lock()
			if reactivate != nil {
				k_lock.Unlock()
				return reactivate(nctx, vals)
			}
			f.Fulfill(eval, vals.Car)
			reactivate = make_reactivation(idx)
			k_lock.Unlock()
			k.Fn = reactivate
			nctx.K = nil
			return false
		}
		return &k
	}

	arglist[count-1].Cdr = VNil
	for i := 0; i < count-1; i++ {
		arglist[i].Cdr = &(arglist[i+1])
	}
	for i := 0; i < count; i++ {
		switch a := args.Car.(type) {
		case *VPair:
			f := new(Future)
			arglist[i].Car = f
			go eval(&Tail{a, ctx.Env, make_argk(i, f)}, true)
		case VSym:
			arglist[i].Car = ctx.Env.Get(a)
		default:
			arglist[i].Car = args.Car
		}
		args = args.Cdr.(*VPair)
	}
	d_lock.Lock()
	done = true
	d_lock.Unlock()
	for context, _ := range blocked {
		delete(blocked, context)
		go eval(context, false)
	}
	return internal.Call(eval, &sctx, &(arglist[0]))
}

func lazyWrapper(internal Callable, eval Evaller, ctx *Tail, args *VPair) bool {
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
	arglist := make([]VPair, count, count)

	make_reactivation := func(idx int) *Continuation {
		return &Continuation{"Thunk", func(ctx *Tail, vals *VPair) bool {
			*ctx = sctx
			nargs, slot := copy_arglist(idx, &(arglist[0]))
			slot.Car = vals.Car
			return internal.Call(eval, ctx, nargs)
		}}
	}

	arglist[count-1].Cdr = VNil
	for i := 0; i < count-1; i++ {
		arglist[i].Cdr = &(arglist[i+1])
	}
	for i := 0; i < count; i++ {
		switch a := args.Car.(type) {
		case *VPair:
			arglist[i].Car = MakeThunk(a, ctx.Env, make_reactivation(i))
		case VSym:
			arglist[i].Car = ctx.Env.Get(a)
		default:
			arglist[i].Car = args.Car
		}
		args = args.Cdr.(*VPair)
	}
	return internal.Call(eval, &sctx, &(arglist[0]))
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
			nargs, slot := copy_arglist(idx, &(arglist[0]))
			slot.Car = vals.Car
			return internal.Call(eval, nctx, nargs)
		}
		return func(nctx *Tail, vals *VPair) bool {
			f_lock.RLock()
			if finished {
				f_lock.RUnlock()
				return reassign(nctx, vals)
			}
			f_lock.RUnlock()
			blocked[&Tail{vals.Car, sctx.Env, &Continuation{"ArgK", reassign}}] = struct{}{}
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
	argset := make(map[int]VValue)
	for i := 0; i < count-1; i++ {
		arglist[i].Cdr = &(arglist[i+1])
		argset[i] = args.Car
		args = args.Cdr.(*VPair)
	}
	arglist[count-1].Cdr = VNil
	argset[count-1] = args.Car

	for i, a := range argset {
		switch at := a.(type) {
		case *VPair:
			eval(&Tail{at, ctx.Env, make_argk(i)}, true)
		case VSym:
			eval(&Tail{at, ctx.Env, make_argk(i)}, true)
		default:
			arglist[i].Car = a
		}
	}
	f_lock.Lock()
	finished = true
	f_lock.Unlock()
	for context, _ := range blocked {
		delete(blocked, context)
		go eval(context, false)
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
	} else if arg, ok := x.Car.(Callable); ok {
		ctx.Expr = arg
	} else {
		panic("Can't unwrap non-callable")
	}
	return false
}
