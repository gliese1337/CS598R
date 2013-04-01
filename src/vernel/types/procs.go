package types

import "fmt"

type Continuation struct {
	Name string
	Fn   func(*Tail, *VPair) bool
	Refs []VValue
}

func (k *Continuation) Call(_ Evaller, ctx *Tail, x *VPair) bool {
	return k.Fn(ctx, x)
}

func (k *Continuation) GetSize(seen map[VValue]struct{}) int {
	if _, ok := seen[k]; k == nil || ok {
		return 0
	}
	seen[k] = struct{}{}
	total := 1
	for _, v := range k.Refs {
		total += v.GetSize(seen)
	}
	return total
}

func (k *Continuation) String() string {
	return fmt.Sprintf("<cont:%s>", k.Name)
}

var Top = &Continuation{
	"Top",
	func(ctx *Tail, args *VPair) bool {
		ctx.Expr, ctx.K = args.Car, nil
		return false
	},
	nil,
}

type NativeFn struct {
	Name string
	Fn   func(Evaller, *Tail, *VPair) bool
}

func (nfn *NativeFn) Call(eval Evaller, ctx *Tail, x *VPair) bool {
	return nfn.Fn(eval, ctx, x)
}

func (nfn *NativeFn) GetSize(seen map[VValue]struct{}) int {
	if _, ok := seen[nfn]; nfn == nil || ok {
		return 0
	}
	seen[nfn] = struct{}{}
	return 1
}

func (nfn *NativeFn) String() string {
	return fmt.Sprintf("<native:%s>", nfn.Name)
}

func match_args(fs VValue, a *VPair) map[VSym]VValue {
	m := make(map[VSym]VValue)
	switch f := fs.(type) {
	case *VPair:
		for f != nil {
			if a == nil {
				panic("Too few arguments")
			}
			s, ok := f.Car.(VSym)
			if !ok {
				panic("Cannot bind to non-symbol")
			}
			if string(s) != "##" {
				m[s] = a.Car
			}

			ap, ok := a.Cdr.(*VPair)
			switch fp := f.Cdr.(type) {
			case *VPair:
				f, a = fp, ap
			case VSym:
				if ok {
					if string(fp) != "##" {
						m[fp] = ap
					}
					f = nil
				}
			}
		}
	case VSym:
		if string(f) != "##" {
			m[f] = a
		}
	default:
		panic("Invalid formals!")
	}
	return m
}

type Combiner struct {
	Cenv    *Environment
	Formals VValue
	Dsym    VSym
	Body    *VPair
}

func (c *Combiner) Call(_ Evaller, ctx *Tail, args *VPair) bool {
	arg_map := match_args(c.Formals, args)
	if c.Body == nil {
		ctx.Expr = VNil
		return false
	}
	if string(c.Dsym) != "##" {
		arg_map[c.Dsym] = WrapEnv(ctx.Env)
	}
	senv, sk := NewEnv(c.Cenv, arg_map), ctx.K
	var eloop func(*Tail, *VPair) bool
	eloop = func(kctx *Tail, body *VPair) bool {
		next_expr, ok := body.Cdr.(*VPair)
		if !ok {
			panic("Invalid Function Body")
		}
		if next_expr == nil {
			kctx.K = &Continuation{"seq", func(nctx *Tail, va *VPair) bool {
				nctx.Expr, nctx.K = va.Car, sk
				return false
			}, []VValue{sk}}
		} else {
			kctx.K = &Continuation{"seq", func(nctx *Tail, va *VPair) bool {
				return eloop(nctx, next_expr)
			}, []VValue{next_expr}}
		}
		kctx.Expr, kctx.Env = body.Car, senv
		return true
	}
	return eloop(ctx, c.Body)
}

func (c *Combiner) GetSize(seen map[VValue]struct{}) int {
	if _, ok := seen[c]; c == nil || ok {
		return 0
	}
	seen[c] = struct{}{}
	return 1 + c.Formals.GetSize(seen) + c.Body.GetSize(seen) + c.Cenv.GetSize(seen)
}

func (c *Combiner) String() string {
	return "<combiner>"
}

type Applicative struct {
	Wrapper  func(Callable, Evaller, *Tail, *VPair) bool
	Internal Callable
}

func (a *Applicative) Call(eval Evaller, ctx *Tail, args *VPair) bool {
	return a.Wrapper(a.Internal, eval, ctx, args)
}

func (a *Applicative) GetSize(seen map[VValue]struct{}) int {
	if _, ok := seen[a]; a == nil || ok {
		return 0
	}
	seen[a] = struct{}{}
	return 1 + a.Internal.GetSize(seen)
}

func (a *Applicative) String() string {
	if _, ok := a.Internal.(*Environment); ok {
		return "<applicative: Env>"
	}
	return fmt.Sprintf("<applicative: %s>", a.Internal)
}
