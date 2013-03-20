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

func vpanic(_ Evaller, _ *Tail, x *VPair) bool {
	if x == nil {
		panic("Runtime Error")
	}
	panic(x.Car)
}

func timer(eval Evaller, ctx *Tail, x *VPair) bool {
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
	//TODO: Auto-sequence
	sk := ctx.K
	ctx.Expr = expr.Car
	ctx.K = &Continuation{"TimerK", func(nctx *Tail, vals *VPair) bool {
		fmt.Printf("%s ran in %v.\n", label, time.Since(start))
		nctx.Expr = vals.Car
		nctx.K = sk
		return false
	}}
	return true
}

func qand(_ Evaller, ctx *Tail, x *VPair) bool {
	if x == nil {
		panic("No arguments to qand")
	}
	var ok bool
	for ; x != nil; x, ok = x.Cdr.(*VPair) {
		b, ok := x.Car.(VBool)
		if !ok {
			panic("Non-boolean argument to qand")
		}
		if !bool(b) {
			ctx.Expr = b
			return false
		}
	}
	if !ok {
		panic("Invalid Argument List")
	}
	ctx.Expr = VBool(true)
	return false
}

func qor(_ Evaller, ctx *Tail, x *VPair) bool {
	if x == nil {
		panic("No arguments to qor")
	}
	var ok bool
	for ; x != nil; x, ok = x.Cdr.(*VPair) {
		b, ok := x.Car.(VBool)
		if !ok {
			panic("Non-boolean argument to qor")
		}
		if bool(b) {
			ctx.Expr = b
			return false
		}
	}
	if !ok {
		panic("Invalid Argument List")
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

func qeq(_ Evaller, ctx *Tail, x *VPair) bool {
	if x == nil {
		panic("No arguments to qeq")
	}
	var cdr *VPair
	var ok bool
	for ; ; x = cdr {
		cdr, ok = x.Cdr.(*VPair)
		if ok {
			if cdr == nil {
				ctx.Expr = VBool(true)
				return false
			}
			if !equal(x.Car, cdr.Car) {
				ctx.Expr = VBool(false)
				return false
			}
		} else {
			ctx.Expr = VBool(x.Car == cdr)
			return false
		}
	}
	return false
}

func qmul(_ Evaller, ctx *Tail, x *VPair) bool {
	if x == nil {
		panic("No arguments to qmul")
	}
	var ret float64 = 1
	var ok bool
	for ; x != nil; x, ok = x.Cdr.(*VPair) {
		b, ok := x.Car.(VNum)
		if !ok {
			panic("Non-numeric argument to qmul")
		}
		ret *= float64(b)
	}
	if !ok {
		panic("Invalid Argument List")
	}
	ctx.Expr = VNum(ret)
	return false
}

func qdiv(_ Evaller, ctx *Tail, x *VPair) bool {
	if x == nil {
		panic("No arguments to qdiv")
	}
	first, ok := x.Car.(VNum)
	if !ok {
		panic("Non-numeric argument to qdiv")
	}
	ret := float64(first)
	x, ok = x.Cdr.(*VPair)
	for ; x != nil; x, ok = x.Cdr.(*VPair) {
		b, ok := x.Car.(VNum)
		if !ok {
			panic("Non-numeric argument to qdiv")
		}
		ret /= float64(b)
	}
	if !ok {
		panic("Invalid Argument List")
	}
	ctx.Expr = VNum(ret)
	return false
}

func qadd(_ Evaller, ctx *Tail, x *VPair) bool {
	if x == nil {
		panic("No arguments to qadd")
	}
	var ret float64 = 0
	var ok bool
	for ; x != nil; x, ok = x.Cdr.(*VPair) {
		b, ok := x.Car.(VNum)
		if !ok {
			panic("Non-numeric argument to qadd")
		}
		ret += float64(b)
	}
	if !ok {
		panic("Invalid Argument List")
	}
	ctx.Expr = VNum(ret)
	return false
}

func qsub(_ Evaller, ctx *Tail, x *VPair) bool {
	if x == nil {
		panic("No arguments to qsub")
	}
	first, ok := x.Car.(VNum)
	if !ok {
		panic("Non-numeric argument to qsub")
	}
	ret := float64(first)
	x, ok = x.Cdr.(*VPair)
	for ; x != nil; x, ok = x.Cdr.(*VPair) {
		b, ok := x.Car.(VNum)
		if !ok {
			panic("Non-numeric argument to qsub")
		}
		ret -= float64(b)
	}
	if !ok {
		panic("Invalid Argument List")
	}
	ctx.Expr = VNum(ret)
	return false
}

func qless(_ Evaller, ctx *Tail, x *VPair) bool {
	if x == nil {
		panic("No arguments to qless")
	}
	first, ok := x.Car.(VNum)
	if !ok {
		panic("Non-numeric argument to qless")
	}
	last := float64(first)
	x, ok = x.Cdr.(*VPair)
	for ; x != nil; x, ok = x.Cdr.(*VPair) {
		b, ok := x.Car.(VNum)
		if !ok {
			panic("Non-numeric argument to qless")
		}
		next := float64(b)
		if last >= next {
			ctx.Expr = VBool(false)
			return false
		}
		last = next
	}
	ctx.Expr = VBool(true)
	return false
}

func qlesseq(_ Evaller, ctx *Tail, x *VPair) bool {
	if x == nil {
		panic("No arguments to qlesseq")
	}
	first, ok := x.Car.(VNum)
	if !ok {
		panic("Non-numeric argument to qlesseq")
	}
	last := float64(first)
	x, ok = x.Cdr.(*VPair)
	for ; x != nil; x, ok = x.Cdr.(*VPair) {
		b, ok := x.Car.(VNum)
		if !ok {
			panic("Non-numeric argument to qlesseq")
		}
		next := float64(b)
		if last > next {
			ctx.Expr = VBool(false)
			return false
		}
		last = next
	}
	ctx.Expr = VBool(true)
	return false
}

func qgreater(_ Evaller, ctx *Tail, x *VPair) bool {
	if x == nil {
		panic("No arguments to qgreater")
	}
	first, ok := x.Car.(VNum)
	if !ok {
		panic("Non-numeric argument to qgreater")
	}
	last := float64(first)
	x, ok = x.Cdr.(*VPair)
	for ; x != nil; x, ok = x.Cdr.(*VPair) {
		b, ok := x.Car.(VNum)
		if !ok {
			panic("Non-numeric argument to qlesseq")
		}
		next := float64(b)
		if last <= next {
			ctx.Expr = VBool(false)
			return false
		}
		last = next
	}
	ctx.Expr = VBool(true)
	return false
}

func qgreatereq(_ Evaller, ctx *Tail, x *VPair) bool {
	if x == nil {
		panic("No arguments to qgreatereq")
	}
	first, ok := x.Car.(VNum)
	if !ok {
		panic("Non-numeric argument to qgreatereq")
	}
	last := float64(first)
	x, ok = x.Cdr.(*VPair)
	for ; x != nil; x, ok = x.Cdr.(*VPair) {
		b, ok := x.Car.(VNum)
		if !ok {
			panic("Non-numeric argument to qlesseq")
		}
		next := float64(b)
		if last < next {
			ctx.Expr = VBool(false)
			return false
		}
		last = next
	}
	ctx.Expr = VBool(true)
	return false
}

func qisbool(_ Evaller, ctx *Tail, x *VPair) bool {
	_, ok := x.Car.(VBool)
	ctx.Expr = VBool(ok)
	return false
}
func qisnum(_ Evaller, ctx *Tail, x *VPair) bool {
	_, ok := x.Car.(VNum)
	ctx.Expr = VBool(ok)
	return false
}
func qispair(_ Evaller, ctx *Tail, x *VPair) bool {
	_, ok := x.Car.(*VPair)
	ctx.Expr = VBool(ok)
	return false
}
func qisstr(_ Evaller, ctx *Tail, x *VPair) bool {
	_, ok := x.Car.(VStr)
	ctx.Expr = VBool(ok)
	return false
}
func qissym(_ Evaller, ctx *Tail, x *VPair) bool {
	_, ok := x.Car.(VSym)
	ctx.Expr = VBool(ok)
	return false
}
func qisproc(_ Evaller, ctx *Tail, x *VPair) bool {
	_, ok := x.Car.(Callable)
	ctx.Expr = VBool(ok)
	return false
}
func qisapp(_ Evaller, ctx *Tail, x *VPair) bool {
	_, ok := x.Car.(*Applicative)
	ctx.Expr = VBool(ok)
	return false
}
func qislazy(_ Evaller, ctx *Tail, x *VPair) bool {
	_, ok := x.Car.(Deferred)
	ctx.Expr = VBool(ok)
	return false
}
func qisthunk(_ Evaller, ctx *Tail, x *VPair) bool {
	_, ok := x.Car.(*Thunk)
	ctx.Expr = VBool(ok)
	return false
}
func qisfuture(_ Evaller, ctx *Tail, x *VPair) bool {
	_, ok := x.Car.(*Future)
	ctx.Expr = VBool(ok)
	return false
}

func vdefer(_ Evaller, ctx *Tail, x *VPair) bool {
	ctx.Expr = MakeThunk(x.Car, ctx.Env, ctx.K)
	return false
}

func spawn(eval Evaller, ctx *Tail, x *VPair) bool {
	f := MakeFuture(x.Car, ctx.Env, ctx.K)
	f.Run(eval)
	ctx.Expr = f
	return false
}

func qstrict(eval Evaller, ctx *Tail, x *VPair) bool {
	if d, ok := x.Car.(Deferred); ok {
		return d.Strict(eval, ctx)
	}
	ctx.Expr = x.Car
	return false
}

func qread(_ Evaller, ctx *Tail, x *VPair) bool {
	if x == nil {
		ctx.Expr = VNil
		return false
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

func use(eval Evaller, ctx *Tail, x *VPair) bool {
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
	return true
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

func load(eval Evaller, ctx *Tail, x *VPair) bool {
	env := GetBuiltins()
	loader(eval, env, x, "Non-string argument to load")
	ctx.Expr = WrapEnv(env)
	return false
}

func qimport(eval Evaller, ctx *Tail, x *VPair) bool {
	loader(eval, ctx.Env, x, "Non-string argument to import")
	ctx.Expr = VNil
	return false
}

func abort(_ Evaller, ctx *Tail, x *VPair) bool {
	ctx.K = Top
	if x == nil {
		ctx.Expr = VNil
		return false
	}
	ctx.Expr = x.Car
	return true
}

func bindcc(_ Evaller, ctx *Tail, x *VPair) bool {
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
	ctx.Expr, ctx.Env = body.Car, NewEnv(ctx.Env, map[VSym]VValue{
		k_sym: &Applicative{func(_ Callable, _ Evaller, nctx *Tail, args *VPair) bool {
			if args == nil {
				return sk.Fn(nctx, VNil)
			}
			*nctx = Tail{args.Car, senv, sk}
			return true
		}, sk},
	})
	return true
}

func qcons(_ Evaller, ctx *Tail, x *VPair) bool {
	if x == nil {
		panic("No Arguments to cons")
	}
	if cdr, ok := x.Cdr.(*VPair); ok {
		if cdr == nil {
			panic("Too few arguments to cons")
		}
		ctx.Expr = &VPair{x.Car, cdr.Car}
		return false
	}
	panic(fmt.Sprintf("Invalid Arguments to cons: %v", x))
}

func qcar(_ Evaller, ctx *Tail, x *VPair) bool {
	if x == nil {
		panic("No Argument to car")
	}
	if arg, ok := x.Car.(*VPair); ok {
		if arg == nil {
			panic("Empty List Passed to car")
		}
		ctx.Expr = arg.Car
		return false
	}
	panic(fmt.Sprintf("Invalid Argument to car: %v", x))
}

func qcdr(_ Evaller, ctx *Tail, x *VPair) bool {
	if x == nil {
		panic("No Argument to cdr")
	}
	if arg, ok := x.Car.(*VPair); ok {
		if arg == nil {
			panic("Empty List Passed to cdr")
		}
		ctx.Expr = arg.Cdr
		return false
	}
	panic(fmt.Sprintf("Invalid Argument to cdr: %v", x))
}

func last(_ Evaller, ctx *Tail, x *VPair) bool {
	var nx *VPair
	var ok bool
	for ; x != nil; x = nx {
		if nx, ok = x.Cdr.(*VPair); ok {
			if nx == nil {
				ctx.Expr = x.Car
				return false
			}
		} else {
			panic("Invalid Argument List")
		}
	}
	ctx.Expr = VNil
	return false
}

func qlist(_ Evaller, ctx *Tail, x *VPair) bool {
	ctx.Expr = x
	return false
}

func def(_ Evaller, ctx *Tail, x *VPair) bool {
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
	var val VValue
	if rest == nil {
		val = VNil
	} else {
		val = rest.Car
	}
	sk, senv := ctx.K, ctx.Env
	ctx.Expr, ctx.K = val, &Continuation{
		"def",
		func(nctx *Tail, args *VPair) bool {
			senv.Set(sym, args.Car)
			*nctx = Tail{args.Car, senv, sk}
			return false
		},
	}
	return true
}

func acc_syms(sset *(map[string]struct{}), arg interface{}) {
	for arg != nil {
		switch a := arg.(type) {
		case *VPair:
			acc_syms(sset, a.Car)
			arg = a.Cdr
		case VSym:
			(*sset)[string(a)] = struct{}{}
			arg = nil
		default:
			arg = nil
		}
	}
}

func unique(_ Evaller, ctx *Tail, x *VPair) bool {
	var sset map[string]struct{}
	acc_syms(&sset, x)
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

func qprint(_ Evaller, ctx *Tail, x *VPair) bool {
	(&p_lock).Lock()
	for x != nil {
		fmt.Printf("%v", x.Car)
		x, _ = x.Cdr.(*VPair)
	}
	(&p_lock).Unlock()
	ctx.Expr = VNil
	return false
}
