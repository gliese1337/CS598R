package eval

import (
	"fmt"
	. "vernel/types"
)

func proc_k(sctx Tail, args *VPair) *Continuation {
	return &Continuation{"call", func(nctx *Tail, p ...VValue) bool {
		nctx.K = &Continuation{"call",func(kctx *Tail, f ..VValue) bool {
			if proc, ok := f[0].(Callable); ok {
				*kctx = sctx
				return proc.Call(eval_loop, nctx, args)
			}
			panic("Non-callable in function position")
		}}
		return p[0].Strict(nctx)
	}}

}

func eval_loop(state *Tail, evaluate bool) {
	for state.K != nil {
		if evaluate {
			switch xt := state.Expr.(type) {
			case VSym:
				state.Expr = state.Env.Get(xt)
			case *VPair:
				if xt != nil {
					state.Expr, state.K = xt.Car, proc_k(*state, xt.Cdr)
					continue
				}
			}
		}
		evaluate = state.K.Fn(state, state.Expr)
	}
}

func Eval(x VValue, env *Environment, cb func(*VPair)) {
	eval_loop(&Tail{x, env, &Continuation{
		"Top",
		func(ctx *Tail, vals *VPair) bool {
			cb(vals)
			ctx.K = nil
			return false
		},
	}}, true)
}
