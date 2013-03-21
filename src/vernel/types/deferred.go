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
	}}
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

func (f *Future) Run(eval Evaller) {
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
		for context, _ := range f.blocked {
			context.Expr = v
			go eval(context, false)
		}
		f.blocked = nil
		ctx.K = nil
		return false
	}}
	go eval(&Tail{f.expr, f.env, k}, true)
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
			}}
			return d.Strict(eval, ctx)
		}
		ctx.Expr = f.result
		return false
	}
	f.lock.Unlock()
	f.blocked[&Tail{nil, ctx.Env, ctx.K}] = struct{}{}
	ctx.K = nil
	return false
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
			}}
			return d.Strict(eval, ctx)
		}
		ctx.Expr = t.result
		return false
	}
	if t.executing {
		t.blocked[&Tail{nil, ctx.Env, ctx.K}] = struct{}{}
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
		for context, _ := range t.blocked {
			context.Expr = v
			go eval(context, false)
		}
		t.blocked = nil
		nctx.Expr, nctx.Env, nctx.K = v, senv, sk
		return false
	}}
	return true
}

func (t *Thunk) String() string {
	if t.fulfilled {
		return fmt.Sprintf("%v", t.result)
	}
	return "<thunk>"
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
			}}
			return d.Strict(eval, ctx)
		}
		ctx.Expr = e.result
		return false
	}
	if e.executing {
		e.blocked[&Tail{nil, ctx.Env, ctx.K}] = struct{}{}
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

			for context, _ := range e.blocked {
				context.Expr = v
				go eval(context, false)
			}
			e.blocked = nil
			kctx.Expr, kctx.Env, kctx.K = v, senv, sk
			return false
		}}
		nctx.Expr, nctx.Env = v, e.env
		return true
	}}
	return e.d.Strict(eval, ctx)
}

func (e *EvalDefer) String() string {
	if e.fulfilled {
		return fmt.Sprintf("%v", e.result)
	}
	return "<thunk>"
}
