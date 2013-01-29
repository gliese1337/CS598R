package eval

import (
	. "vernel/types"
)

func proc_k(env *Environment, k *Continuation, args *VPair) *Continuation {
	return &Continuation{
		func(p *VPair) *Tail {
			if p != nil {
				if proc, ok := p.Car.(Callable); ok {
					return proc.Call(Eval, env, k, args)
				}
			}
			panic("Non-callable in function position")
		},
	}

}

func Eval(x interface{}, env *Environment, k *Continuation) interface{} {
	for k != nil {
		switch xt := x.(type) {
		case VSym:
			x = env.Get(xt)
		case *VPair:
			if xt != nil {
				arglist, ok := xt.Cdr.(*VPair)
				if !ok {
					panic("Non-list in argument position")
				}
				x, k = xt.Car, proc_k(env, k, arglist)
				continue
			}
		}
		tail := k.Fn(&VPair{x, VNil})
		x, env, k = tail.Expr, tail.Env, tail.K
	}
	return x
}
