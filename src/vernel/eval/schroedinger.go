package eval

import (
	. "vernel/types"
)

func Eval(x interface{}, env *Environment) interface{} {
start:
	switch xt := x.(type) {
	case VSym:
		return env.Get(xt)
	case *VPair: //(proc exp*)
		var rec bool
		if xt == nil {
			return VNil
		}
		arglist, ok := xt.Cdr.(*VPair)
		if !ok {
			panic("Non-list in argument position")
		}
		proc, ok := Eval(xt.Car, env).(Callable)
		if !ok {
			panic("Non-callable in procedure position")
		}
		x, env, rec = proc.Call(Eval, env, arglist)
		if rec {
			goto start
		}
	}
	return x
}
