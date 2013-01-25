package lib

import . "vernel/types"

var Standard = NewEnv(
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
	})
