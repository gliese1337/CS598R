package eval

import (
        . "vernel/types"
)

func Eval(x interface{}, env *Environment) interface{} {
	start:
    switch xt := x.(type) {
    case VSym:
        x = env.Get(xt)
    case *VPair: //(proc exp*)
		arglist, ok := xt.Cdr.(*VPair)
		if !ok { panic("Non-list in argument position") }
		val := Eval(xt.Car, env)
        proc, ok := val.(Callable)
		if !ok { panic("Non-callable in procedure position") }
    	x = proc.Call(Eval, env, arglist)
		if tail, ok := x.(*Tail); ok {
			x, env = tail.Expr, tail.Env
			goto start
		}
	}
	return x                                                                 
} 
