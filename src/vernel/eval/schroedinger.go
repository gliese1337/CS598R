package eval

import (
	"fmt"
	. "vernel/types"
)

func check_args(proc Callable, args *VPair) []VValue {
	var arglist []VValue
	argc, variadic := proc.Arity()
	if variadic {
		arglist = make([]VValue,argc+1)
	} else {
		arglist = make([]VValue,argc)
	}
	carg := args
	for i := 0; i < argc; i++ {
		if carg == nil {
			panic(fmt.Sprintf("%d arguments given to %d-arity function",i,argc))
		}
		arglist[i] = carg.Car
		carg = carg.Cdr.(*VPair)
	}
	if variadic {
		arglist[argc] = carg
	} else if carg != nil {
		panic("Too many arguments to non-variadic function")
	}
	return arglist
}

func proc_k(sctx Tail, args *VPair) *Continuation {
	return &Continuation{
		Name: "call",
		Argc: 1,
		Variadic: false,
		Fn: func(nctx *Tail, p ...VValue) bool {
			if proc, ok := p[0].(Callable); ok {
				*nctx = sctx
				arglist := check_args(proc,args)
				return proc.Call(eval_loop, nctx, arglist...)
			}
			panic("Non-callable in function position")
		},
	}

}

func eval_loop(state *Tail, evaluate bool) {
	for state.K != nil {
		if evaluate {
			switch xt := state.Expr.(type) {
			case VSym:
				state.Expr = state.Env.Get(xt)
			case *VPair:
				if xt != nil {
					arglist, ok := xt.Cdr.(*VPair)
					if !ok {
						panic(fmt.Sprintf("Non-list \"%s\" in argument position", xt.Cdr))
					}
					state.Expr, state.K = xt.Car, proc_k(*state, arglist)
					continue
				}
			}
		}
		evaluate = state.K.Call(nil,state,state.Expr)
	}
}

func Eval(x VValue, env *Environment, cb func(...VValue)) {
	eval_loop(&Tail{x, env, &Continuation{
		Name: "Top",
		Argc: 1,
		Variadic: false,
		Fn: func(ctx *Tail, vals ...VValue) bool {
			cb(vals...)
			ctx.K = nil
			return false
		},
	}}, true)
}
