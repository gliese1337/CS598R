package types

import (
	"bytes"
	"fmt"
	"strconv"
)

type VValue interface {
	Strict(*Tail) bool
	String() string
}

type Tail struct {
	Expr VValue
	Env  *Environment
	K    *Continuation
}

func (t *Tail) Return(x *VPair) bool {
	return t.K.Fn(t, x)
}

type Evaller func(*Tail,bool)

type Callable interface {
	Call(Evaller, *Tail, *VPair) bool
}

type VSym string
func (v VSym) Strict(ctx *Tail) bool {
	ctx.Expr = v
	return false
}
func (v VSym) String() string {
	return string(v)
}

type VStr string
func (v VStr) Strict(ctx *Tail) bool {
	ctx.Expr = v
	return false
}
func (v VStr) String() string {
	return string(v)
}

type VNum float64
func (v VNum) Strict(ctx *Tail) bool {
	ctx.Expr = v
	return false
}
func (v VNum) String() string {
	return strconv.FormatFloat(float64(v), 'g', -1, 64)
}

type VBool bool
func (v VBool) Strict(ctx *Tail) bool {
	ctx.Expr = v
	return false
}
func (v VBool) String() string {
	if v {
		return "#t"
	}
	return "#f"
}

func (v VBool) Call(eval Evaller, ctx *Tail, args *VPair) bool {
	if args == nil {
		ctx.Expr = VNil
	} else {
		cdr, ok := args.Cdr.(*VPair)
		if !ok || cdr == nil {
			panic(fmt.Sprintf("Invalid Arguments to Branch: %v", args))
		}
		if bool(v) {
			ctx.Expr = args.Car
		} else {
			ctx.Expr = cdr.Car
		}
	}
	return true
}

type VPair struct {
	Car interface{}
	Cdr interface{}
}
func (v *VPair) Strict(ctx *Tail) bool {
	ctx.Expr = v
	return false
}
func (v *VPair) String() string {
	if v == nil {
		return "()"
	}
	var buf bytes.Buffer
	buf.WriteRune('(')
write_rest:
	buf.WriteString(fmt.Sprintf("%s", v.Car))
	tail, ok := v.Cdr.(*VPair)
	if ok {
		if tail == nil {
			buf.WriteRune(')')
		} else {
			buf.WriteRune(' ')
			v = tail
			goto write_rest
		}
	} else {
		buf.WriteString(" . ")
		buf.WriteString(fmt.Sprintf("%s", v.Cdr))
		buf.WriteRune(')')
	}
	return buf.String()
}

var VNil *VPair = nil

type Future struct {
	channel chan VValue
	lock sync.Mutex
	result VValue
	fulfilled bool
	blocked map[*Tail]struct{}
}
func MakeFuture() *Future {
	return &Future{
		lock: new(sync.RWMutex),
		blocked: make(map[*Tail]struct{})
		result: nil
		fulfilled: false
	}
}
func (f *Future Fulfill(v VValue) {
	f.lock.Lock()
	if f.fulfilled {
		f.lock.Unlock()
		panic("Cannot fulfill future more than once.")
	}
	f.fulfilled = true
	f.lock.Unlock()
	f.result = v
	for context, _ := range blocked {
		delete(blocked,context)
		go eval(context,false)
	}
	
}
func (f *Future) Strict(ctx *Tail) bool {
	f.lock.RLock()
	if f.fulfilled {
		f.lock.RUnlock()
		ctx.Expr = f.result
		return false
	}
	f.lock.RUnlock
	blocked[&Tail{f.result, ctx.Env, ctx.K}] = struct{}{}
	ctx.K = nil
	return false
}
func (f *Future) String() string {
	if f.fulfilled {
		return fmt.Sprintf("%v",f.Result)
	}
	return "<future>"
}