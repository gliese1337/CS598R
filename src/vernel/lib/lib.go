package lib

import . "vernel/types"

func GetBuiltins() *Environment {
	return NewEnv(
		nil,
		map[VSym]interface{}{
			VSym("def"):      &NativeFn{"define", def},
			VSym("qcons"):    &NativeFn{"qcons", qcons},
			VSym("qcar"):     &NativeFn{"qcar", qcar},
			VSym("qcdr"):     &NativeFn{"qcdr", qcdr},
			VSym("vau"):      &NativeFn{"vau", vau},
			VSym("bind/cc"):  &NativeFn{"bind/cc", bindcc},
			VSym("qunwrap"):  &NativeFn{"qunwrap", qunwrap},
			VSym("wrap/rtl"): wrap_gen(rtlWrapper),
			VSym("wrap/snc"): wrap_gen(syncWrapper),
			VSym("wrap"):     wrap_gen(basicWrapper),
			VSym("qprint"):   &NativeFn{"qprint", qprint},
			VSym("last"):     &NativeFn{"last", last},
			VSym("qlist"):    &NativeFn{"qlist", qlist},
			VSym("use"):      &NativeFn{"use", use},
			VSym("load"):     &NativeFn{"load", load},
			VSym("import"):   &NativeFn{"import", qimport},
			VSym("qread"):    &NativeFn{"qread", qread},
			VSym("qeq?"):     &NativeFn{"qeq", qeq},
			VSym("qand"):     &NativeFn{"qand", qand},
			VSym("qor"):      &NativeFn{"qor", qor},
			VSym("'*"):       &NativeFn{"qmul", qmul},
			VSym("'/"):       &NativeFn{"qdiv", qdiv},
			VSym("'+"):       &NativeFn{"qadd", qadd},
			VSym("'-"):       &NativeFn{"qsub", qsub},
			VSym("'<"):       &NativeFn{"qless", qless},
			VSym("'<="):      &NativeFn{"qlesseq", qlesseq},
			VSym("'>"):       &NativeFn{"qgreater", qgreater},
			VSym("'>="):      &NativeFn{"qgreatereq", qgreatereq},
			VSym("qbool?"):   &NativeFn{"qbool?", qisbool},
			VSym("qnum?"):    &NativeFn{"qnum?", qisnum},
			VSym("qstr?"):    &NativeFn{"qstr?", qisstr},
			VSym("qsym?"):    &NativeFn{"qsym?", qissym},
			VSym("qpair?"):   &NativeFn{"qpair?", qispair},
			VSym("timer"):    &NativeFn{"timer", timer},
			VSym("panic"):    &NativeFn{"panic", vpanic},
			VSym("abort"):    &NativeFn{"abort", abort},
			VSym("qunique"):  &NativeFn{"qunique", unique},
		})
}
