package types

import "fmt"
import "sync"

type Deferred interface {
	VValue
	Strict(Evaller, *Tail) bool
}

func force(eval Evaller, x VValue, ctx *Tail) bool {
	if d, ok := x.(Deferred); ok {
		return d.Strict(eval, ctx)
	}
	ctx.Expr = x
	return false
}

func strict_block(eval Evaller, ctx *Tail) {
	k := ctx.K
	ctx.K = &Continuation{"s block", func(ctx *Tail, vals *VPair) bool {
		if d, ok := vals.Car.(Deferred); ok {
			return d.Strict(eval, ctx)
		}
		return k.Fn(ctx, vals)
	}, []VValue{k}}
}

type Future struct {
	lock      sync.Mutex
	expr      VValue
	env       *Environment
	k         *Continuation
	result    VValue
	fulfilled bool
	blocked   map[*Tail]struct{}
}

func MakeFuture(expr VValue, env *Environment, k *Continuation) *Future {
	return &Future{expr: expr, env: env, k: k, blocked: make(map[*Tail]struct{})}
}

func (f *Future) Run(eval Evaller, time int) {
	k := &Continuation{"FTop", func(ctx *Tail, vals *VPair) bool {
		v := vals.Car
		f.lock.Lock()
		if f.fulfilled {
			f.lock.Unlock()
			ctx.Expr, ctx.Env, ctx.K = v, f.env, f.k
			return false
		}
		f.result = v
		f.fulfilled = true
		f.lock.Unlock()
		ntime := ctx.Time
		for context, _ := range f.blocked {
			context.Expr = v
			go eval(context, ntime, false)
		}
		f.blocked = nil
		ctx.K = nil
		return false
	}, []VValue{f}}
	go eval(&Tail{f.expr, f.env, k, time}, time, true)
	f.expr = nil
}

func (f *Future) Strict(eval Evaller, ctx *Tail) bool {
	strict_block(eval, ctx)
	f.lock.Lock()
	if f.fulfilled {
		f.lock.Unlock()
		if d, ok := f.result.(Deferred); ok {
			sk := ctx.K
			ctx.K = &Continuation{"replace", func(nctx *Tail, vals *VPair) bool {
				f.result = vals.Car
				return sk.Fn(nctx, vals)
			}, []VValue{f, sk}}
			return d.Strict(eval, ctx)
		}
		ctx.Expr = f.result
		return false
	}
	f.lock.Unlock()
	f.blocked[&Tail{nil, ctx.Env, ctx.K, ctx.Time}] = struct{}{}
	ctx.K = nil
	return false
}

func (f *Future) GetSize(seen map[VValue]struct{}) int {
	if _, ok := seen[f]; f == nil || ok {
		return 0
	}
	seen[f] = struct{}{}
	f.lock.Lock()
	if f.fulfilled {
		f.lock.Unlock()
		return 1 + f.result.GetSize(seen)
	}
	f.lock.Unlock()
	return 1 + f.expr.GetSize(seen) + f.env.GetSize(seen) + f.k.GetSize(seen)
}

func (f *Future) String() string {
	if f.fulfilled {
		return fmt.Sprintf("%v", f.result)
	}
	return "<future>"
}

type Thunk struct {
	lock      sync.Mutex
	result    VValue
	fulfilled bool
	executing bool
	expr      VValue
	env       *Environment
	k         *Continuation
	blocked   map[*Tail]struct{}
}

func MakeThunk(expr VValue, env *Environment, k *Continuation) *Thunk {
	return &Thunk{expr: expr, env: env, k: k, blocked: make(map[*Tail]struct{})}
}

func (t *Thunk) Strict(eval Evaller, ctx *Tail) bool {
	strict_block(eval, ctx)
	t.lock.Lock()
	if t.fulfilled {
		t.lock.Unlock()
		if d, ok := t.result.(Deferred); ok {
			sk := ctx.K
			ctx.K = &Continuation{"replace", func(nctx *Tail, vals *VPair) bool {
				t.result = vals.Car
				return sk.Fn(nctx, vals)
			}, []VValue{t, sk}}
			return d.Strict(eval, ctx)
		}
		ctx.Expr = t.result
		return false
	}
	if t.executing {
		t.blocked[&Tail{nil, ctx.Env, ctx.K, ctx.Time}] = struct{}{}
		ctx.K = nil
		return false
	}
	t.executing = true
	t.lock.Unlock()
	senv, sk := ctx.Env, ctx.K
	ctx.Expr, ctx.Env = t.expr, t.env
	ctx.K = &Continuation{"thunk", func(nctx *Tail, vals *VPair) bool {
		v := vals.Car
		t.lock.Lock()
		if t.fulfilled {
			t.lock.Unlock()
			nctx.Expr, nctx.Env, nctx.K = v, t.env, t.k
			return false
		}
		t.fulfilled = true
		t.result = v
		t.lock.Unlock()
		ntime := nctx.Time
		for context, _ := range t.blocked {
			context.Expr = v
			go eval(context, ntime, false)
		}
		t.blocked = nil
		nctx.Expr, nctx.Env, nctx.K = v, senv, sk
		return false
	}, []VValue{senv, sk, t}}
	return true
}

func (t *Thunk) GetSize(seen map[VValue]struct{}) int {
	if _, ok := seen[t]; t == nil || ok {
		return 0
	}
	seen[t] = struct{}{}
	t.lock.Lock()
	if t.fulfilled {
		t.lock.Unlock()
		return 1 + t.result.GetSize(seen)
	}
	t.lock.Unlock()
	return 1 + t.expr.GetSize(seen) + t.env.GetSize(seen) + t.k.GetSize(seen)
}

func (t *Thunk) String() string {
	if t.fulfilled {
		return fmt.Sprintf("%v", t.result)
	}
	return fmt.Sprintf("<thunk:%v>", t.expr)
}

type EvalDefer struct {
	lock      sync.Mutex
	result    VValue
	fulfilled bool
	executing bool
	d         Deferred
	env       *Environment
	k         *Continuation
	blocked   map[*Tail]struct{}
}

func MakeEvalDefer(d Deferred, env *Environment, k *Continuation) *EvalDefer {
	return &EvalDefer{d: d, env: env, k: k, blocked: make(map[*Tail]struct{})}
}

func (e *EvalDefer) Strict(eval Evaller, ctx *Tail) bool {
	strict_block(eval, ctx)
	e.lock.Lock()
	if e.fulfilled {
		e.lock.Unlock()
		if d, ok := e.result.(Deferred); ok {
			sk := ctx.K
			ctx.K = &Continuation{"replace", func(nctx *Tail, vals *VPair) bool {
				e.result = vals.Car
				return sk.Fn(nctx, vals)
			}, []VValue{e, sk}}
			return d.Strict(eval, ctx)
		}
		ctx.Expr = e.result
		return false
	}
	if e.executing {
		e.blocked[&Tail{nil, ctx.Env, ctx.K, ctx.Time}] = struct{}{}
		ctx.K = nil
		return false
	}
	e.executing = true
	e.lock.Unlock()
	senv, sk := ctx.Env, ctx.K
	ctx.K = &Continuation{"strict", func(nctx *Tail, vals *VPair) bool {
		//receives the resulting of stricting the contained deferred
		v := vals.Car
		e.lock.Lock()
		if e.fulfilled {
			e.lock.Unlock()
			nctx.Expr, nctx.Env, nctx.K = v, e.env, e.k
			return false
		}
		e.d = nil
		e.lock.Unlock()

		//need to now evaluate the result of stricting
		nctx.K = &Continuation{"evaldefer", func(kctx *Tail, vals *VPair) bool {
			v := vals.Car
			e.lock.Lock()
			if e.fulfilled {
				e.lock.Unlock()
				kctx.Expr = v
				return false
			}
			e.fulfilled = true
			e.result = v
			e.lock.Unlock()

			ntime := kctx.Time
			for context, _ := range e.blocked {
				context.Expr = v
				go eval(context, ntime, false)
			}
			e.blocked = nil
			kctx.Expr, kctx.Env, kctx.K = v, senv, sk
			return false
		}, []VValue{e, senv, sk}}
		nctx.Expr, nctx.Env = v, e.env
		return true
	}, []VValue{e, senv, sk}}
	return e.d.Strict(eval, ctx)
}

func (e *EvalDefer) GetSize(seen map[VValue]struct{}) int {
	if _, ok := seen[e]; e == nil || ok {
		return 0
	}
	seen[e] = struct{}{}
	e.lock.Lock()
	if e.fulfilled {
		e.lock.Unlock()
		return 1 + e.result.GetSize(seen)
	}
	e.lock.Unlock()
	return 1 + e.d.GetSize(seen) + e.env.GetSize(seen) + e.k.GetSize(seen)
}

func (e *EvalDefer) String() string {
	if e.fulfilled {
		return fmt.Sprintf("%v", e.result)
	}
	return "<thunk>"
}
