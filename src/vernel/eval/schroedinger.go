package eval

import (
	"fmt"
	//"runtime"
	. "vernel/types"
)

func proc_k(sctx Tail, args *VPair) *Continuation {
	return &Continuation{
		"call",
		func(nctx *Tail, p *VPair) {
			if p != nil {
				if proc, ok := p.Car.(Callable); ok {
					//fmt.Printf("Calling %v with arguments %v\n", proc, args)
					*nctx = sctx
					proc.Call(Eval, nctx, args)
					return
				}
			}
			panic("Non-callable in function position")
		},
	}

}

func Eval(x interface{}, env *Environment, k *Continuation) interface{} {
	state := Tail{x, env, k}
	//_, f, l, _ := runtime.Caller(1)
	//fmt.Printf("Eval called from %v:%v\n", f, l)
	for state.K != nil {
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
				state.Expr, state.K = xt.Car, proc_k(state, arglist)
				continue
			}
		}
		(&state).Return(&VPair{state.Expr, VNil})
	}
	return state.Expr
}
