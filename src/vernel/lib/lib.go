package lib

import . "vernel/types"

func GetBuiltins() *Environment {
	return NewEnv(
		nil,
		map[VSym]VValue{
			VSym("def"):      &NativeFn{"define", 2, false, def},
			VSym("qcons"):    &NativeFn{"qcons", 2, false, qcons},
			VSym("qcar"):     &NativeFn{"qcar", 1, false, qcar},
			VSym("qcdr"):     &NativeFn{"qcdr", 1, false, qcdr},
			VSym("vau"):      &NativeFn{"vau", 2, true, vau},
			VSym("bind/cc"):  &NativeFn{"bind/cc", 1, true, bindcc},
			VSym("qunwrap"):  &NativeFn{"qunwrap", 1, false, qunwrap},
			VSym("wrap/rtl"): wrap_gen(rtlWrapper),
			VSym("wrap/snc"): wrap_gen(syncWrapper),
			VSym("wrap"):     wrap_gen(basicWrapper),
			VSym("qprint"):   &NativeFn{"qprint", 0, true, qprint},
			VSym("last"):     &NativeFn{"last", 0, true, last},
			VSym("qlist"):    &NativeFn{"qlist", 0, true, qlist},
			VSym("use"):      &NativeFn{"use", 1, true, use},
			VSym("load"):     &NativeFn{"load", 0, true, load},
			VSym("import"):   &NativeFn{"import", 0, true, qimport},
			VSym("qread"):    &NativeFn{"qread", 0, true, qread},
			VSym("qeq?"):     &NativeFn{"qeq", 0, true, qeq},
			VSym("qand"):     &NativeFn{"qand", 0, true, qand},
			VSym("qor"):      &NativeFn{"qor", 0, true, qor},
			VSym("'*"):       &NativeFn{"qmul", 0, true, qmul},
			VSym("'/"):       &NativeFn{"qdiv", 0, true, qdiv},
			VSym("'+"):       &NativeFn{"qadd", 0, true, qadd},
			VSym("'-"):       &NativeFn{"qsub", 0, true, qsub},
			VSym("'<"):       &NativeFn{"qless", 0, true, qless},
			VSym("'<="):      &NativeFn{"qlesseq", 0, true, qlesseq},
			VSym("'>"):       &NativeFn{"qgreater", 0, true, qgreater},
			VSym("'>="):      &NativeFn{"qgreatereq", 0, true, qgreatereq},
			VSym("qbool?"):   &NativeFn{"qbool?", 1, false, qisbool},
			VSym("qnum?"):    &NativeFn{"qnum?", 1, false, qisnum},
			VSym("qstr?"):    &NativeFn{"qstr?", 1, false, qisstr},
			VSym("qsym?"):    &NativeFn{"qsym?", 1, false, qissym},
			VSym("qpair?"):   &NativeFn{"qpair?", 1, false, qispair},
			VSym("timer"):    &NativeFn{"timer", 2, true, timer},
			VSym("panic"):    &NativeFn{"panic", 1, false, vpanic},
			VSym("abort"):    &NativeFn{"abort", 1, false, abort},
			VSym("qunique"):  &NativeFn{"qunique", 0, true, unique},
		})
}
