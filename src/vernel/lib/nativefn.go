package lib

import (
	. "vernel/types"
	"vernel/parser"
	"fmt"
	"os"
	"bufio"
	"time"
)

func timer(eval Evaller, env *Environment, k *Continuation, x *VPair) *Tail {
	if x == nil {
		panic("No arguments to timer.")
	}
	label, ok := x.Car.(VStr)
	if !ok {
		panic("Invalid timer label.")
	}
	expr, ok := x.Cdr.(*VPair)
	if !ok || expr == nil {
		panic("Invalid timer expression.")
	}
	start := time.Now()
	val := eval(expr.Car,env,Top)
	fmt.Printf("%s ran in %v.\n", label, time.Now().Sub(start))
	return k.Fn(&VPair{val,VNil})
}

func qand(_ Evaller, _ *Environment, k *Continuation, x *VPair) *Tail {
	if x == nil { panic("No arguments to qand") }
	for x != nil {
		b, ok := x.Car.(VBool)
		if !ok {
			panic("Non-boolean argument to qand")
		}
		if !bool(b) {
			return k.Fn(&VPair{b,VNil})
		}
		cdr, ok := x.Cdr.(*VPair)
		if ok {
			x = cdr
		} else {
			x = &VPair{cdr,VNil}
		}
	}
	return k.Fn(&VPair{VBool(true),VNil})
}

func qor(_ Evaller, _ *Environment, k *Continuation, x *VPair) *Tail {
	if x == nil { panic("No arguments to qor") }
	for x != nil {
		b, ok := x.Car.(VBool)
		if !ok {
			panic("Non-boolean argument to qor")
		}
		if bool(b) {
			return k.Fn(&VPair{b,VNil})
		}
		cdr, ok := x.Cdr.(*VPair)
		if ok {
			x = cdr
		} else {
			x = &VPair{cdr,VNil}
		}
	}
	return k.Fn(&VPair{VBool(false),VNil})
}

func qeq(_ Evaller, _ *Environment, k *Continuation, x *VPair) *Tail {
	if x == nil { panic("No arguments to qeq") }
	var ret bool
	for {
		cdr, ok := x.Cdr.(*VPair)
		if ok {
			if cdr == nil {
				ret = true
				break
			}
			if x.Car != cdr.Car {
				ret = false
				break
			}
			x = cdr
		} else {
			ret = x.Car == cdr
			break
		}
	}
	return k.Fn(&VPair{VBool(ret),VNil})
}

func qread(_ Evaller, _ *Environment, k *Continuation, x *VPair) *Tail {
	if x == nil {
		return k.Fn(&VPair{VNil,VNil})
	}
	inchan := make(chan rune)
	go func() {
		for x != nil {
			vstr, ok := x.Car.(VStr)
			if !ok {
				panic("Non-string argument to read")
			}
			for _, r := range string(vstr) {
				inchan <- r
			}
		}
		close(inchan)
	}()
	var rootpair = VPair{nil,nil}
	curpair := &rootpair
	for expr := range parser.Parse(inchan) {
		nextpair := &VPair{expr,VNil}
		curpair.Cdr = nextpair
		curpair = nextpair
	}
	return k.Fn(&VPair{rootpair.Cdr,VNil})
}

func load_env(eval Evaller, env *Environment, fname string) {
	file, err := os.Open(fname)
	if err != nil {
		panic("Error opening file.")
	}
	defer file.Close()
	
	inchan := make(chan rune)
	go func() {
		freader := bufio.NewReader(file)
	loop:
		if r, _, err := freader.ReadRune(); err == nil {
			inchan <- r
			goto loop
		}
		close(inchan)
	}()
	for expr := range parser.Parse(inchan) {
		eval(expr, env, Top)
	}
}

func use(eval Evaller, _ *Environment, k *Continuation, x *VPair) *Tail {
	if x == nil { panic("No arguments to use") }
	vstr, ok := x.Car.(VStr)
	if !ok { panic("Non-string argument to use") }
	body, ok := x.Cdr.(*VPair)
	if !ok { panic("Missing body expression in use") }
	env := GetBuiltins()
	load_env(eval, env, string(vstr))
	return &Tail{body.Car,env,k}
}

func loader(eval Evaller, env *Environment, x *VPair, pstr string) {
	for x != nil {
		vstr, ok := x.Car.(VStr)
		if !ok { panic(pstr) }
		load_env(eval, env, string(vstr))
		x, ok = x.Cdr.(*VPair)
	}
}

func load(eval Evaller, _ *Environment, k *Continuation, x *VPair) *Tail {
	env := GetBuiltins()
	loader(eval, env, x, "Non-string argument to load")
	return k.Fn(&VPair{WrapEnv(env),VNil})
}

func qimport(eval Evaller, env *Environment, k *Continuation, x *VPair) *Tail {
	loader(eval, env, x, "Non-string argument to import")
	return k.Fn(&VPair{VNil,VNil})
}

func bindcc(_ Evaller, senv *Environment, k *Continuation, x *VPair) *Tail {
	if x == nil {
		panic("No arguments to bind/cc")
	}
	k_sym, ok := x.Car.(VSym)
	if !ok {
		panic("Cannot bind to non-symbol")
	}
	body, ok := x.Cdr.(*VPair)
	if !ok {
		panic("No body provided to bind/cc")
	}
	return &Tail{body.Car, NewEnv(senv, map[VSym]interface{}{
		k_sym: &Applicative{func(_ Callable, _ Evaller, cenv *Environment, _ *Continuation, args *VPair) *Tail {
			if args == nil {
				return k.Fn(VNil)
			}
			return &Tail{args.Car, cenv, k}
		}, k},
	}), k}
}

func qcons(_ Evaller, _ *Environment, k *Continuation, x *VPair) *Tail {
	if x == nil {
		panic("No Arguments to cons")
	}
	if cdr, ok := x.Cdr.(*VPair); ok {
		if cdr == nil {
			panic("Too few arguments to cons")
		}
		return k.Fn(&VPair{&VPair{x.Car, cdr.Car}, VNil})
	}
	panic("Invalid Arguments to cons")
}

func qcar(_ Evaller, _ *Environment, k *Continuation, x *VPair) *Tail {
	if x == nil {
		panic("No Argument to car")
	}
	if arg, ok := x.Car.(*VPair); ok {
		if arg == nil {
			panic("Empty List Passed to car")
		}
		return k.Fn(&VPair{arg.Car, VNil})
	}
	panic("Invalid Argument to car")
}

func qcdr(_ Evaller, _ *Environment, k *Continuation, x *VPair) *Tail {
	if x == nil {
		panic("No Argument to cdr")
	}
	if arg, ok := x.Car.(*VPair); ok {
		if arg == nil {
			panic("Empty List Passed to cdr")
		}
		return k.Fn(&VPair{arg.Cdr, VNil})
	}
	panic("Invalid Argument to cdr")
}

func last(_ Evaller, _ *Environment, k *Continuation, x *VPair) *Tail {
	var nx *VPair
	var ok bool
	for ; x != nil; x = nx {
		if nx, ok = x.Cdr.(*VPair); ok {
			if nx == nil {
				return k.Fn(&VPair{x, VNil})
			}
		} else {
			panic("Invalid Argument List")
		}
	}
	return k.Fn(VNil)
}

func qlist(_ Evaller, _ *Environment, k *Continuation, x *VPair) *Tail {
	return k.Fn(&VPair{x, VNil})
}

func def(_ Evaller, env *Environment, k *Continuation, x *VPair) *Tail {
	if x == nil {
		panic("No Arguments to def")
	}
	sym, ok := x.Car.(VSym)
	if !ok {
		panic("Cannot define non-symbol")
	}
	rest, ok := x.Cdr.(*VPair)
	if !ok {
		panic("Non-list argument to def")
	}
	var val interface{}
	if rest == nil {
		val = VNil
	} else {
		val = rest.Car
	}
	return &Tail{val, env, &Continuation{
		func(args *VPair) *Tail {
			env.Set(sym, args.Car)
			return k.Fn(&VPair{args.Car, VNil})
		},
	}}
}

func qprint(_ Evaller, env *Environment, k *Continuation, x *VPair) *Tail {
	var val interface{}
	if x == nil {
		val = VNil
	} else {
		val = x.Car
	}
	fmt.Printf("%s", val)
	return &Tail{val, env, k}
}
