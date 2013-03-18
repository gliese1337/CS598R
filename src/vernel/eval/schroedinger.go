package eval

import (
	"fmt"
	. "vernel/types"
)

func proc_k(sctx Tail, args *VPair) *Continuation {
	var callk func(*Tail, *VPair) bool
	callk = func(nctx *Tail, p *VPair) bool {
		if p != nil {
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
				return proc.Call(eval_loop, nctx, args)
			}
		}
		panic("Non-callable in function position")
	}
	return &Continuation{"call", callk}
}

func eval_loop(state *Tail, evaluate bool) {
	for state.K != nil {
		if evaluate {
			//TODO: wrap in thunk to propagate laziness
			if d, ok := state.Expr.(Deferred); ok {
				evaluate = d.Strict(eval_loop, state)
				continue
			}
			switch xt := state.Expr.(type) {
			case VSym:
				state.Expr = state.Env.Get(xt)
			case *VPair:
				if xt != nil {
					arglist, ok := xt.Cdr.(*VPair)
					if !ok {
						panic(fmt.Sprintf("Non-list \"%s\" in argument position", xt.Cdr))
					}
					//fmt.Printf("Evaluating procedure expression: %v\n", xt.Car)
					state.Expr, state.K = xt.Car, proc_k(*state, arglist)
					continue
				}
			}
		}
		evaluate = state.K.Fn(state, &VPair{state.Expr, VNil})
	}
}

func Eval(x VValue, env *Environment, k *Continuation) {
	eval_loop(&Tail{x, env, k}, true)
}
