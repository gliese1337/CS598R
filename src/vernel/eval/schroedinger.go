package eval

import . "vernel/types"

func strict_args(proc Callable, ctx *Tail, args VValue) bool {
	var arglist VPair
	var argk func(*Tail, VValue) bool
	senv, sk := ctx.Env, ctx.K
	last_slot := &arglist
	argk = func(ctx *Tail, arg VValue) bool {
	start:
		if slot, ok := arg.(*VPair); ok {
			last_slot.Cdr = slot
			if slot == VNil {
				ctx.Env, ctx.K = senv, sk
				return proc.Call(eval_loop, ctx, arglist.Cdr.(*VPair))
			}
			last_slot, arg = slot, slot.Cdr
			goto start
		}
		if d, ok := arg.(Deferred); ok {
			ctx.K = &Continuation{"strict_args", func(nctx *Tail, vals *VPair) bool {
				return argk(nctx, vals.Car)
			}}
			return d.Strict(eval_loop, ctx)
		}
		panic("Invalid Argument List")
	}
	return argk(ctx, args)
}

func proc_k(sctx Tail, args VValue) *Continuation {
	var callk func(*Tail, *VPair) bool
	callk = func(nctx *Tail, p *VPair) bool {
		if d, ok := p.Car.(Deferred); ok {
			senv, sk := nctx.Env, nctx.K
			nctx.K = &Continuation{"StrictK", func(kctx *Tail, vals *VPair) bool {
				kctx.Env, kctx.K = senv, sk
				return callk(kctx, vals)
			}}
			return d.Strict(eval_loop, nctx)
		}
		if proc, ok := p.Car.(Callable); ok {
			*nctx = sctx
			return strict_args(proc, nctx, args)
		}
		panic("Non-callable in function position")
	}
	return &Continuation{"call", callk}
}

func eval_loop(state *Tail, evaluate bool) {
	for state.K != nil {
		if !evaluate {
			goto finish
		}
		//TODO: wrap in thunk to propagate laziness
		if d, ok := state.Expr.(Deferred); ok {
			evaluate = d.Strict(eval_loop, state)
			continue
		}
		switch xt := state.Expr.(type) {
		case VSym:
			state.Expr = state.Env.Get(xt)
		case *VPair:
			if xt == nil {
				goto finish
			}
			state.Expr, state.K = xt.Car, proc_k(*state, xt.Cdr)
			continue
		}
	finish:
		evaluate = state.K.Fn(state, &VPair{state.Expr, VNil})
	}
}

func Eval(x VValue, env *Environment, k *Continuation) {
	eval_loop(&Tail{x, env, k}, true)
}
