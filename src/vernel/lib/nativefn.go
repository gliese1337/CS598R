package lib

import (
	"bufio"
	"fmt"
	"os"
	"time"
	"vernel/parser"
	. "vernel/types"
)

func vpanic(_ Evaller, _ *Tail, x *VPair) {
	if x == nil {
		panic("Runtime Error")
	}
	panic(x.Car)
}

func timer(eval Evaller, ctx *Tail, x *VPair) {
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
	val := eval(expr.Car, ctx.Env, Top)
	fmt.Printf("%s ran in %v.\n", label, time.Since(start))
	ctx.Return(&VPair{val, VNil})
}

func qand(_ Evaller, ctx *Tail, x *VPair) {
	if x == nil {
		panic("No arguments to qand")
	}
	for x != nil {
		b, ok := x.Car.(VBool)
		if !ok {
			panic("Non-boolean argument to qand")
		}
		if !bool(b) {
			ctx.Expr = b
			return
		}
		cdr, ok := x.Cdr.(*VPair)
		if ok {
			x = cdr
		} else {
			x = &VPair{cdr, VNil}
		}
	}
	ctx.Expr = VBool(true)
}

func qor(_ Evaller, ctx *Tail, x *VPair) {
	if x == nil {
		panic("No arguments to qor")
	}
	for x != nil {
		b, ok := x.Car.(VBool)
		if !ok {
			panic("Non-boolean argument to qor")
		}
		if bool(b) {
			ctx.Expr = b
			return
		}
		cdr, ok := x.Cdr.(*VPair)
		if ok {
			x = cdr
		} else {
			x = &VPair{cdr, VNil}
		}
	}
	ctx.Expr = VBool(false)
}

func equal(a interface{}, b interface{}) bool {
	if at, _ := a.(*VPair); at != nil {
		if bt, _ := b.(*VPair); bt != nil {
			return equal(at.Car, bt.Car) && equal(at.Cdr, bt.Cdr)
		}
		return false
	}
	return a == b
}

func qeq(_ Evaller, ctx *Tail, x *VPair) {
	if x == nil {
		panic("No arguments to qeq")
	}
	for {
		cdr, ok := x.Cdr.(*VPair)
		if ok {
			if cdr == nil {
				ctx.Expr = VBool(true)
				break
			}
			if !equal(x.Car, cdr.Car) {
				ctx.Expr = VBool(false)
				break
			}
			x = cdr
		} else {
			ctx.Expr = VBool(x.Car == cdr)
			break
		}
	}
}

func qmul(_ Evaller, ctx *Tail, x *VPair) {
	if x == nil {
		panic("No arguments to qmul")
	}
	var ret float64 = 1
	for x != nil {
		b, ok := x.Car.(VNum)
		if !ok {
			panic("Non-numeric argument to qmul")
		}
		ret *= float64(b)
		cdr, ok := x.Cdr.(*VPair)
		if ok {
			x = cdr
		} else {
			x = &VPair{cdr, VNil}
		}
	}
	ctx.Expr = VNum(ret)
}

func qdiv(_ Evaller, ctx *Tail, x *VPair) {
	if x == nil {
		panic("No arguments to qdiv")
	}
	first, ok := x.Car.(VNum)
	if !ok {
		panic("Non-numeric argument to qdiv")
	}
	ret := float64(first)
	x, ok = x.Cdr.(*VPair)
	if !ok {
		x = &VPair{x.Cdr, VNil}
	}
	for x != nil {
		b, ok := x.Car.(VNum)
		if !ok {
			panic("Non-numeric argument to qmul")
		}
		ret /= float64(b)
		cdr, ok := x.Cdr.(*VPair)
		if ok {
			x = cdr
		} else {
			x = &VPair{cdr, VNil}
		}
	}
	ctx.Expr = VNum(ret)
}

func qadd(_ Evaller, ctx *Tail, x *VPair) {
	if x == nil {
		panic("No arguments to qadd")
	}
	var ret float64 = 0
	for x != nil {
		b, ok := x.Car.(VNum)
		if !ok {
			panic("Non-numeric argument to qadd")
		}
		ret += float64(b)
		cdr, ok := x.Cdr.(*VPair)
		if ok {
			x = cdr
		} else {
			x = &VPair{cdr, VNil}
		}
	}
	ctx.Expr = VNum(ret)
}

func qsub(_ Evaller, ctx *Tail, x *VPair) {
	if x == nil {
		panic("No arguments to qsub")
	}
	first, ok := x.Car.(VNum)
	if !ok {
		panic("Non-numeric argument to qsub")
	}
	ret := float64(first)
	x, ok = x.Cdr.(*VPair)
	if !ok {
		x = &VPair{x.Cdr, VNil}
	}
	for x != nil {
		b, ok := x.Car.(VNum)
		if !ok {
			panic("Non-numeric argument to qsub")
		}
		ret -= float64(b)
		cdr, ok := x.Cdr.(*VPair)
		if ok {
			x = cdr
		} else {
			x = &VPair{cdr, VNil}
		}
	}
	ctx.Expr = VNum(ret)
}

func qisbool(_ Evaller, ctx *Tail, x *VPair) {
	_, ok := x.Car.(VBool)
	ctx.Expr = VBool(ok)
}
func qisnum(_ Evaller, ctx *Tail, x *VPair) {
	_, ok := x.Car.(VNum)
	ctx.Expr = VBool(ok)
}
func qispair(_ Evaller, ctx *Tail, x *VPair) {
	_, ok := x.Car.(*VPair)
	ctx.Expr = VBool(ok)
}
func qisstr(_ Evaller, ctx *Tail, x *VPair) {
	_, ok := x.Car.(VStr)
	ctx.Expr = VBool(ok)
}
func qissym(_ Evaller, ctx *Tail, x *VPair) {
	_, ok := x.Car.(VSym)
	ctx.Expr = VBool(ok)
}

func qread(_ Evaller, ctx *Tail, x *VPair) {
	if x == nil {
		ctx.Return(&VPair{VNil, VNil})
		return
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
	var rootpair = VPair{nil, nil}
	curpair := &rootpair
	for expr := range parser.Parse(inchan) {
		nextpair := &VPair{expr, VNil}
		curpair.Cdr = nextpair
		curpair = nextpair
	}
	ctx.Return(&VPair{rootpair.Cdr, VNil})
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

func use(eval Evaller, ctx *Tail, x *VPair) {
	if x == nil {
		panic("No arguments to use")
	}
	vstr, ok := x.Car.(VStr)
	if !ok {
		panic("Non-string argument to use")
	}
	body, ok := x.Cdr.(*VPair)
	if !ok {
		panic("Missing body expression in use")
	}
	env := GetBuiltins()
	load_env(eval, env, string(vstr))
	ctx.Expr, ctx.Env = body.Car, env
}

func loader(eval Evaller, env *Environment, x *VPair, pstr string) {
	for x != nil {
		vstr, ok := x.Car.(VStr)
		if !ok {
			panic(pstr)
		}
		load_env(eval, env, string(vstr))
		x, ok = x.Cdr.(*VPair)
	}
}

func load(eval Evaller, ctx *Tail, x *VPair) {
	env := GetBuiltins()
	loader(eval, env, x, "Non-string argument to load")
	ctx.Expr = WrapEnv(env)
}

func qimport(eval Evaller, ctx *Tail, x *VPair) {
	loader(eval, ctx.Env, x, "Non-string argument to import")
	ctx.Return(&VPair{VNil, VNil})
}

func bindcc(_ Evaller, ctx *Tail, x *VPair) {
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
	sk, senv := ctx.K, ctx.Env
	ctx.Expr, ctx.Env = body.Car, NewEnv(ctx.Env, map[VSym]interface{}{
		k_sym: &Applicative{func(_ Callable, _ Evaller, nctx *Tail, args *VPair) {
			if args == nil {
				sk.Fn(nctx, VNil)
			} else {
				*nctx = Tail{args.Car, senv, sk}
			}
		}, sk},
	})
}

func qcons(_ Evaller, ctx *Tail, x *VPair) {
	if x == nil {
		panic("No Arguments to cons")
	}
	if cdr, ok := x.Cdr.(*VPair); ok {
		if cdr == nil {
			panic("Too few arguments to cons")
		}
		ctx.Return(&VPair{&VPair{x.Car, cdr.Car}, VNil})
		return
	}
	panic(fmt.Sprintf("Invalid Arguments to cons: %v", x))
}

func qcar(_ Evaller, ctx *Tail, x *VPair) {
	if x == nil {
		panic("No Argument to car")
	}
	if arg, ok := x.Car.(*VPair); ok {
		if arg == nil {
			panic("Empty List Passed to car")
		}
		ctx.Return(&VPair{arg.Car, VNil})
		return
	}
	panic(fmt.Sprintf("Invalid Argument to car: %v", x))
}

func qcdr(_ Evaller, ctx *Tail, x *VPair) {
	if x == nil {
		panic("No Argument to cdr")
	}
	if arg, ok := x.Car.(*VPair); ok {
		if arg == nil {
			panic("Empty List Passed to cdr")
		}
		ctx.Return(&VPair{arg.Cdr, VNil})
		return
	}
	panic(fmt.Sprintf("Invalid Argument to cdr: %v", x))
}

func last(_ Evaller, ctx *Tail, x *VPair) {
	var nx *VPair
	var ok bool
	for ; x != nil; x = nx {
		if nx, ok = x.Cdr.(*VPair); ok {
			if nx == nil {
				ctx.Return(&VPair{x.Car, VNil})
				return
			}
		} else {
			panic("Invalid Argument List")
		}
	}
	ctx.Return(VNil)
}

func qlist(_ Evaller, ctx *Tail, x *VPair) {
	ctx.Return(&VPair{x, VNil})
}

func def(_ Evaller, ctx *Tail, x *VPair) {
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
	sk, senv := ctx.K, ctx.Env
	ctx.Expr, ctx.K = val, &Continuation{
		"def",
		func(nctx *Tail, args *VPair) {
			senv.Set(sym, args.Car)
			nctx.Env, nctx.K = senv, sk
			nctx.Return(&VPair{args.Car, VNil})
		},
	}
}

func qprint(_ Evaller, ctx *Tail, x *VPair) {
	for x != nil {
		fmt.Printf("%v", x.Car)
		x, _ = x.Cdr.(*VPair)
	}
	ctx.Return(&VPair{VNil, VNil})
}
