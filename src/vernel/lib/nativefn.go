package lib

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"
	"vernel/parser"
	. "vernel/types"
)

func vpanic(_ Evaller, _ *Tail, x ...VValue) bool {
	panic(x[0])
}

func timer(eval Evaller, ctx *Tail, x ...VValue) bool {
	label, ok := x[0].(VStr)
	if !ok {
		panic("Invalid timer label.")
	}
	start := time.Now()
	//TODO: Auto-sequence
	sk := ctx.K
	ctx.Expr = x[1]
	ctx.K = &Continuation{
		Name: "TimerK",
		Argc: 1,
		Variadic: false,
		Fn: func(nctx *Tail, vals ...VValue) bool {
			fmt.Printf("%s ran in %v.\n", label, time.Since(start))
			nctx.Expr = vals[0]
			nctx.K = sk
			return false
		},
	}
	return true
}

func qand(_ Evaller, ctx *Tail, x ...VValue) bool {
	for _, v := range x {
		b, ok := v.(VBool)
		if !ok {
			panic("Non-boolean argument to qand")
		}
		if !bool(b) {
			ctx.Expr = b
			return false
		}
	}
	ctx.Expr = VBool(true)
	return false
}

func qor(_ Evaller, ctx *Tail, x ...VValue) bool {
	for _, v := range x {
		b, ok := v.(VBool)
		if !ok {
			panic("Non-boolean argument to qor")
		}
		if bool(b) {
			ctx.Expr = b
			return false
		}
	}
	ctx.Expr = VBool(false)
	return false
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

func qeq(_ Evaller, ctx *Tail, x ...VValue) bool {
	if len(x) > 0 {
		first := x[0]
		for _, v := range x[1:] {
			if !equal(first, v) {
				ctx.Expr = VBool(false)
				return false
			}
		}
	}
	ctx.Expr = VBool(true)
	return false
}

func qmul(_ Evaller, ctx *Tail, x ...VValue) bool {
	var ret float64 = 1
	for _, v := range x {
		n, ok := v.(VNum)
		if !ok {
			panic("Non-numeric argument to qmul")
		}
		ret *= float64(n)
	}
	ctx.Expr = VNum(ret)
	return false
}

func qdiv(_ Evaller, ctx *Tail, x ...VValue) bool {
	if len(x) == 0 {
		ctx.Expr = VNum(1)
		return false
	}
	first, ok := x[0].(VNum)
	if !ok {
		panic("Non-numeric argument to qdiv")
	}
	ret := float64(first)
	for _, v := range x[1:] {
		n, ok := v.(VNum)
		if !ok {
			panic("Non-numeric argument to qmul")
		}
		ret /= float64(n)
	}
	ctx.Expr = VNum(ret)
	return false
}

func qadd(_ Evaller, ctx *Tail, x ...VValue) bool {
	var ret float64 = 0
	for _, v := range x {
		n, ok := v.(VNum)
		if !ok {
			panic("Non-numeric argument to qadd")
		}
		ret += float64(n)
	}
	ctx.Expr = VNum(ret)
	return false
}

func qsub(_ Evaller, ctx *Tail, x ...VValue) bool {
	if len(x) == 0 {
		ctx.Expr = VNum(0)
		return false
	}
	first, ok := x[0].(VNum)
	if !ok {
		panic("Non-numeric argument to qsub")
	}
	ret := float64(first)
	for _, v := range x {
		n, ok := v.(VNum)
		if !ok {
			panic("Non-numeric argument to qsub")
		}
		ret -= float64(n)
	}
	ctx.Expr = VNum(ret)
	return false
}

func qless(_ Evaller, ctx *Tail, x ...VValue) bool {
	if len(x) < 2 {
		ctx.Expr = VBool(false)
		return false
	}
	first, ok := x[0].(VNum)
	if !ok {
		panic("Non-numeric argument to qless")
	}
	last := float64(first)
	for _, v := range x {
		n, ok := v.(VNum)
		if !ok {
			panic("Non-numeric argument to qless")
		}
		next := float64(n)
		if last >= next {
			ctx.Expr = VBool(false)
			return false
		}
		last = next
	}
	ctx.Expr = VBool(true)
	return false
}

func qlesseq(_ Evaller, ctx *Tail, x ...VValue) bool {
	if len(x) < 2 {
		ctx.Expr = VBool(true)
		return false
	}
	first, ok := x[0].(VNum)
	if !ok {
		panic("Non-numeric argument to qlesseq")
	}
	last := float64(first)
	for _, v := range x {
		n, ok := v.(VNum)
		if !ok {
			panic("Non-numeric argument to qlesseq")
		}
		next := float64(n)
		if last > next {
			ctx.Expr = VBool(false)
			return false
		}
		last = next
	}
	ctx.Expr = VBool(true)
	return false
}

func qgreater(_ Evaller, ctx *Tail, x ...VValue) bool {
	if len(x) < 2 {
		ctx.Expr = VBool(false)
		return false
	}
	first, ok := x[0].(VNum)
	if !ok {
		panic("Non-numeric argument to qgreater")
	}
	last := float64(first)
	for _, v := range x {
		n, ok := v.(VNum)
		if !ok {
			panic("Non-numeric argument to qgreater")
		}
		next := float64(n)
		if last <= next {
			ctx.Expr = VBool(false)
			return false
		}
		last = next
	}
	ctx.Expr = VBool(true)
	return false
}

func qgreatereq(_ Evaller, ctx *Tail, x ...VValue) bool {
	if len(x) < 2 {
		ctx.Expr = VBool(true)
		return false
	}
	first, ok := x[0].(VNum)
	if !ok {
		panic("Non-numeric argument to qgreatereq")
	}
	last := float64(first)
	for _, v := range x {
		n, ok := v.(VNum)
		if !ok {
			panic("Non-numeric argument to qgreatereq")
		}
		next := float64(n)
		if last < next {
			ctx.Expr = VBool(false)
			return false
		}
		last = next
	}
	ctx.Expr = VBool(true)
	return false

}

func qisbool(_ Evaller, ctx *Tail, x ...VValue) bool {
	_, ok := x[0].(VBool)
	ctx.Expr = VBool(ok)
	return false
}
func qisnum(_ Evaller, ctx *Tail, x ...VValue) bool {
	_, ok := x[0].(VNum)
	ctx.Expr = VBool(ok)
	return false
}
func qispair(_ Evaller, ctx *Tail, x ...VValue) bool {
	_, ok := x[0].(*VPair)
	ctx.Expr = VBool(ok)
	return false
}
func qisstr(_ Evaller, ctx *Tail, x ...VValue) bool {
	_, ok := x[0].(VStr)
	ctx.Expr = VBool(ok)
	return false
}
func qissym(_ Evaller, ctx *Tail, x ...VValue) bool {
	_, ok := x[0].(VSym)
	ctx.Expr = VBool(ok)
	return false
}

func qread(_ Evaller, ctx *Tail, x ...VValue) bool {
	if len(x) == 0 {
		ctx.Expr = VNil
		return false
	}
	inchan := make(chan rune)
	go func() {
		for _, v := range x {
			vstr, ok := v.(VStr)
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
	ctx.Expr = rootpair.Cdr
	return false
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
		eval(&Tail{expr, env, nil}, true)
	}
}

func use(eval Evaller, ctx *Tail, x ...VValue) bool {
	vstr, ok := x[0].(VStr)
	if !ok {
		panic("Non-string argument to use")
	}
	body, ok := x[1].(*VPair)
	if !ok {
		panic("Missing body expression in use")
	}
	env := GetBuiltins()
	load_env(eval, env, string(vstr))
	ctx.Expr, ctx.Env = body.Car, env
	return true
}

func loader(eval Evaller, env *Environment, x []VValue, pstr string) {
	for _, v := range x {
		vstr, ok := v.(VStr)
		if !ok {
			panic(pstr)
		}
		load_env(eval, env, string(vstr))
	}
}

func load(eval Evaller, ctx *Tail, x ...VValue) bool {
	env := GetBuiltins()
	loader(eval, env, x, "Non-string argument to load")
	ctx.Expr = WrapEnv(env)
	return false
}

func qimport(eval Evaller, ctx *Tail, x ...VValue) bool {
	loader(eval, ctx.Env, x, "Non-string argument to import")
	ctx.Expr = VNil
	return false
}

func abort(_ Evaller, ctx *Tail, x ...VValue) bool {
	ctx.K = nil
	ctx.Expr = x[0]
	return true
}

func bindcc(_ Evaller, ctx *Tail, x ...VValue) bool {
	k_sym, ok := x[0].(VSym)
	if !ok {
		panic("Cannot bind to non-symbol")
	}
	body, ok := x[1].(*VPair)
	if !ok || body == nil {
		panic("No body provided to bind/cc")
	}
	sk, senv := ctx.K, ctx.Env
	ctx.Expr, ctx.Env = body.Car, NewEnv(ctx.Env, map[VSym]VValue{
		k_sym: &Applicative{
			Internal: sk,
			Wrapper: func(_ Callable, _ Evaller, nctx *Tail, args ...VValue) bool {
				if args == nil {
					return sk.Fn(nctx, VNil)
				}
				*nctx = Tail{args[0], senv, sk}
				return true
			},
		},
	})
	return true
}

func qcons(_ Evaller, ctx *Tail, x ...VValue) bool {
	ctx.Expr = &VPair{x[0], x[1]}
	return false
}

func qcar(_ Evaller, ctx *Tail, x ...VValue) bool {
	if arg, ok := x[0].(*VPair); ok {
		if arg == nil {
			panic("Empty List Passed to car")
		}
		ctx.Expr = arg.Car
		return false
	}
	panic(fmt.Sprintf("Invalid Argument to car: %v", x[0]))
}

func qcdr(_ Evaller, ctx *Tail, x ...VValue) bool {
	if arg, ok := x[0].(*VPair); ok {
		if arg == nil {
			panic("Empty List Passed to cdr")
		}
		ctx.Expr = arg.Cdr
		return false
	}
	panic(fmt.Sprintf("Invalid Argument to cdr: %v", x[0]))
}

func last(_ Evaller, ctx *Tail, x ...VValue) bool {
	ctx.Expr = x[len(x)-1]
	return false
}

func qlist(_ Evaller, ctx *Tail, x ...VValue) bool {
	ctx.Expr = x[0]
	return false
}

func def(_ Evaller, ctx *Tail, x ...VValue) bool {
	sym, ok := x[0].(VSym)
	if !ok {
		panic("Cannot define non-symbol")
	}
	sk, senv := ctx.K, ctx.Env
	ctx.Expr, ctx.K = x[1], &Continuation{
		Name: "def",
		Argc: 1,
		Variadic: false,
		Fn: func(nctx *Tail, args ...VValue) bool {
			senv.Set(sym, args[0])
			*nctx = Tail{args[0], senv, sk}
			return false
		},
	}
	return true
}

func acc_syms(sset *(map[string]struct{}), arg VValue) {
	for arg != nil {
		switch a := arg.(type) {
		case *VPair:
			acc_syms(sset, a.Car)
			arg = a.Cdr
		case VSym:
			(*sset)[string(a)] = struct{}{}
			return
		default:
			return
		}
	}
}

func unique(_ Evaller, ctx *Tail, x ...VValue) bool {
	var sset map[string]struct{}
	acc_syms(&sset, x[0])
	cntr := 0
gen_str:
	ustr := fmt.Sprintf("u%x", cntr)
	if _, ok := sset[ustr]; ok {
		cntr++
		goto gen_str
	}
	ctx.Expr = VSym(ustr)
	return false
}

var p_lock sync.Mutex

func qprint(_ Evaller, ctx *Tail, x ...VValue) bool {
	(&p_lock).Lock()
	for _, v := range x {
		fmt.Printf("%v", v)
	}
	(&p_lock).Unlock()
	ctx.Expr = VNil
	return false
}
