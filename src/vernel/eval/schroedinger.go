package eval

import "fmt"
import . "vernel/types"
import "vernel/prof"

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
			}, nil}
			return d.Strict(eval_loop, ctx)
		}
		panic(fmt.Sprintf("Invalid Argument List: %v", arglist.Cdr))
	}
	return argk(ctx, args)
}

func proc_k(sctx Tail, args VValue) *Continuation {
	var callk func(*Tail, *VPair) bool
	sexpr, senv, sk := sctx.Expr, sctx.Env, sctx.K
	callk = func(nctx *Tail, p *VPair) bool {
		if d, ok := p.Car.(Deferred); ok {
			nctx.K = &Continuation{"StrictK", callk, []VValue{sexpr, senv, sk}}
			return d.Strict(eval_loop, nctx)
		}
		if proc, ok := p.Car.(Callable); ok {
			nctx.Expr, nctx.Env, nctx.K = sexpr, senv, sk
			return strict_args(proc, nctx, args)
		}
		panic("Non-callable in function position")
	}
	return &Continuation{"call", callk, []VValue{sexpr, senv, sk}}
}

func eval_loop(state *Tail, starttime int, evaluate bool) {
	if state.Time < starttime {
		e := state.Expr
		state.Expr = nil
		for state.Time < starttime {
			prof.Clock(state)
		}
		state.Expr = e
	}
	for state.K != nil {
		prof.Clock(state)
		if !evaluate {
			goto finish
		}
		if d, ok := state.Expr.(Deferred); ok {
			state.Expr = MakeEvalDefer(d, state.Env, state.K)
			goto finish
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
	time := prof.GetTime()
	eval_loop(&Tail{x, env, k, time}, time, true)
}
