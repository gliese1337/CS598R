package lib

import . "vernel/types"

func GetBuiltins() *Environment {
	return NewEnv(
	nil,
	map[VSym]interface{}{
		VSym("def"):      &NativeFn{def},
		VSym("qcons"):    &NativeFn{qcons},
		VSym("qcar"):     &NativeFn{qcar},
		VSym("qcdr"):     &NativeFn{qcdr},
		VSym("vau"):      &NativeFn{vau},
		VSym("bind/cc"):  &NativeFn{bindcc},
		VSym("unwrap"):   &NativeFn{unwrap},
		VSym("wrap/rtl"): wrap_gen(rtlWrapper),
		VSym("qprint"):   &NativeFn{qprint},
		VSym("last"):     &NativeFn{last},
		VSym("qlist"):    &NativeFn{qlist},
		VSym("use"):    &NativeFn{use},
		VSym("load"):    &NativeFn{load},
		VSym("import"):    &NativeFn{qimport},
		VSym("qread"):    &NativeFn{qread},
		VSym("qeq?"):    &NativeFn{qeq},
		VSym("qand"):    &NativeFn{qand},
		VSym("qor"):    &NativeFn{qor},
	})
}
