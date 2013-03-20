package types

import "fmt"
import "sync"

type Deferred interface {
	VValue
	Strict(Evaller, *Tail) bool
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
	k := &Continuation{"ArgK", func(nctx *Tail, vals *VPair) bool {
		v := vals.Car
		f.lock.Lock()
		if f.fulfilled {
			f.lock.Unlock()
			nctx.Expr, nctx.Env, nctx.K = v, f.env, f.k
			return false
		}
		f.fulfilled = true
		f.result = v
		f.lock.Unlock()
		for context, _ := range f.blocked {
			delete(f.blocked, context)
			context.Expr = v
			go eval(context, false)
		}
		nctx.K = nil
		return false
	}}
	go eval(&Tail{f.expr, f.env, k}, true)
}

func (f *Future) Strict(_ Evaller, ctx *Tail) bool {
	f.lock.Lock()
	if f.fulfilled {
		f.lock.Unlock()
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
	t.lock.Lock()
	if t.fulfilled {
		t.lock.Unlock()
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
			delete(t.blocked, context)
			context.Expr = v
			go eval(context, false)
		}
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
