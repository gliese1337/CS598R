package types

import (
	"fmt"
	"sync"
)

type Future struct {
	lock      sync.RWMutex
	result    VValue
	fulfilled bool
	blocked   map[*Tail]struct{}
}

func (f *Future) Fulfill(eval Evaller, v VValue) {
	f.lock.Lock()
	if f.fulfilled {
		f.lock.Unlock()
		panic("Cannot fulfill future more than once.")
	}
	f.fulfilled = true
	f.lock.Unlock()
	f.result = v
	for context, _ := range f.blocked {
		delete(f.blocked, context)
		go eval(&Tail{v, context.Env, context.K}, false)
	}

}
func (f *Future) Strict(ctx *Tail) bool {
	f.lock.RLock()
	if f.fulfilled {
		f.lock.RUnlock()
		return f.result.Strict(ctx)
	}
	f.lock.RUnlock()
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
