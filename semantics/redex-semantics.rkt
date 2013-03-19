#lang racket
(require redex)

(define-language vernel
  [Env (env (X Value) ...)]
  [Binding (X any)]
  [Heap (heap Binding ...)]
  [Wraptype wrap
            wrap-lazy
            wrap-future
            wrap-ltr
            wrap-rtl
            wrap-r6rs]
  [Callwrap call-wrap
            call-lazy
            call-future
            call-ltr
            call-rtl
            call-r6rs]
  [Prog (Env Expr)]
  [Pointer (p X)]
  [boolean #t #f]
  [Expr Value
        (thunk Env Expr)
        (unwrap Expr)
        (Wraptype Expr)
        (call Expr Expr ...)]
  [Op boolean Wraptype unwrap defer spawn force vau + - * > < =]
  [Callable Op Env
            (Wraptype Callable)
            (vau (X ...) Y Expr)]
  [Idem Callable number]
  [Evable X Pointer (l Value ...)]
  [Value Evable Idem]
  [Atomic X Op number]
  [BValue Env
          Pointer
          (l Value ... BValue Value ...)
          (vau (X ...) Y Expr)
          (Wraptype BValue)]
  [FValue Atomic (l FValue ...)]
  [Node Wraptype Callwrap call call-fn thunk unwrap thunk l heap env strict]
  [A hole (Heap CC)] ;answer evaluation
  [C (Heap CC) CC]
  [CC hole
      (Env C) ;basic evaluation
      ((heap Binding_1 ... (X C) Binding_2 ...) any) ;parallel evaluation
      (call-fn C any ...) ;fn evaluation
      ;ordered argument evaluation
      (call-wrap Callable any ... C any ...)
      (call-r6rs Callable any ... C any ...)
      (call-ltr Callable Value ... C Prog ...)
      (call-rtl Callable Prog ... C Value ...)]
  [(X Y) variable-not-otherwise-mentioned])

(define reduce
  (reduction-relation
   vernel
   (--> (in-hole C ((heap Binding_1 ...) ((heap Binding_2 ...) any)))
        (in-hole C ((heap Binding_1 ... Binding_2 ...) any))
        "merge")
   (--> (in-hole C_1 (Env (in-hole C_2 (Heap any))))
        (in-hole C_1 (Heap (Env (in-hole C_2 any))))
        (side-condition (term (contains-no-bind C_2)))
        "lift")
   ;garbage collection
   (--> (in-hole C (Heap FValue))
        (in-hole C FValue))
   (--> (in-hole C (Heap BValue))
        (in-hole C ((trim-heap Heap BValue) BValue))
        (side-condition (term (contains-unused Heap BValue)))
        "GC")
   ;eval cases
   (--> (in-hole C (Env Idem))
        (in-hole C Idem))
   (--> (in-hole C (Env X))
        (in-hole C (lookup Env X)))
   (--> (in-hole C (Env (l Value ...)))
        (in-hole C (Env (call Value ...))))
   (--> (in-hole C (Env Pointer))
        (in-hole C (defer Env (strict Pointer))))
   ;strict eval cases
   (--> (in-hole C (Env (strict Idem)))
        (in-hole C (Env Idem)))
   (--> (in-hole C (Env (strict X)))
        (in-hole C (lookup Env X)))
   (--> (in-hole C (Env (strict (l Value ...))))
        (in-hole C (Env (call Value ...))))
   (--> (in-hole C_1 ((heap Binding_1 ... (X (thunk Env_1 Expr)) Binding_2 ...) (in-hole C_2 (Env_2 (strict (p X))))))
        (in-hole C_1 ((heap Binding_1 ... (X (Env_1 Expr)) Binding_2 ...) (in-hole C_2 (p X))))
        "strict")
   ;substitution
   #;(--> (in-hole C_1 ((heap Binding_1 ... (X (in-hole A BValue)) Binding_2 ...) (in-hole C_2 (p X))))
        (in-hole C_1 ((subst X BValue (heap Binding_1 ... (X (in-hole A BValue)) Binding_2 ...)) (subst X BValue (in-hole C_2 (p X)))))
        "bound subst")
   (--> (in-hole C_1 ((heap Binding_1 ... (X (in-hole A Value)) Binding_2 ...) (in-hole C_2 (p X))))
        (in-hole C_1 ((subst X FValue (heap Binding_1 ... Binding_2 ...)) (subst X Value (in-hole C_2 (p X)))))
        "subst")
   ;application rules
   (--> (in-hole C (Env (call Evable Expr ...)))
        (in-hole C (Env (call-fn (Env Evable) Expr ...)))
        "fn lookup")
   (--> (in-hole C (Env (call-fn Callable Expr ...)))
        (in-hole C (Env (call Callable Expr ...)))
        "apply")
   (--> (in-hole C (Env (call (Wraptype Expr_1) Expr_2 ...)))
        (in-hole C (Env (get-call-type Wraptype Env Expr_1 (Expr_2 ...))))
        "arg eval") ;Argument Evaluation & wrap reduction occur in two steps so that ordering rules can take effect
   (--> (in-hole C (Env (call-r6rs Callable Expr_1 ... (thunk Env Expr_2) Expr_3 ...)))
        (in-hole C (Env (call-r6rs Callable Expr_1 ... (Env Expr_2) Expr_3 ...)))
        "choose")
   (--> (in-hole C (Env (Callwrap Expr Value ...)))
        (in-hole C (Env (call Expr Value ...)))
        "dewrap")
   (--> (in-hole C (Env (call (vau (X ..._1) Y Expr_1) Expr_2 ..._1)))
        (in-hole C ((make-env Env (Y (wrap Env)) (X (value Expr_2)) ...) Expr_1))
        "call")
   (--> (in-hole C (Env (unwrap (Wraptype Expr))))
        (in-hole C (Env Expr))
        "unwrap")
   (--> (in-hole C (Env (call Op Expr ...)))
        (in-hole C (delta Env Op Expr ...))
        "delta")
   (--> (in-hole C (call Env Expr))
        (in-hole C (Env Expr)))))

(define-metafunction vernel
  value : Expr -> Value
  [(value Value) Value]
  [(value (call Expr ...)) (l Expr ...)]
  [(value (unwrap Expr)) (l unwrap Expr)]
  [(value (Wraptype Expr)) (l Wraptype Expr)])

(define-metafunction vernel
  lookup : Env X -> Value
  [(lookup (env) X) ,(error "unbound variable")]
  [(lookup (env Binding_1 ... (X Value) Binding_2 ...) X) Value])

(define-metafunction vernel
  contains-no-bind : C -> any
  [(contains-no-bind hole) #t]
  [(contains-no-bind (Env C)) #f]
  [(contains-no-bind (Heap CC)) #f]
  [(contains-no-bind ((heap Binding_1 ... (X C) Binding_2 ...) any))
   (contains-no-bind C)]
  [(contains-no-bind (call-fn C any ...))
   (contains-no-bind C)]
  [(contains-no-bind (Callwrap Callable any ... C any ...))
   (contains-no-bind C)])

(define-metafunction vernel
  subst : X Value any -> any
  [(subst X Value Atomic) Atomic]
  [(subst X Value (p X)) Value]
  [(subst X Value (p Y)) (p Y)]
  [(subst X Value (Y any)) (Y (subst X Value any))]
  [(subst X Value (Node any ...)) (Node (subst X Value any) ...)]
  [(subst X Value (Heap any)) ((subst X Value Heap) (subst X Value any))]
  [(subst X Value (Env any)) ((subst X Value Env) (subst X Value any))]
  [(subst Y Value (vau (X_1 ...) X_2 Expr))
   (vau (X_1 ...) X_2 (subst Y Value Expr))])

(define-metafunction vernel
  [(make-env (env (X_1 Value_1) ...) (X_2 Value_2) ...)
   (env (X_2 Value_2) ... (X_1 Value_1) ...)])

(define-metafunction vernel
  [(build-heap call-lazy Env Expr_1
               (heap Binding ...)
               (args Pointer ...)
               Expr_2 Expr_3 ...)
   ,(let [(s (gensym))]
      (term (build-heap call-lazy Env Expr_1
                        (heap Binding ... (,s (thunk Env Expr_2)))
                        (args Pointer ... (p ,s))
                        Expr_3 ...)))]
  [(build-heap call-future Env Expr_1
               (heap Binding ...)
               (args Pointer ...)
               Expr_2 Expr_3 ...)
   ,(let [(s (gensym))]
      (term (build-heap call-future Env Expr_1
                        (heap Binding ... (,s (Env Expr_2)))
                        (args Pointer ... (p ,s))
                        Expr_3 ...)))]
  [(build-heap Callwrap Env Expr_1 Heap (args Pointer ...))
   (Heap (Callwrap Expr_1 Pointer ...))])

(define-metafunction vernel
  [(trim-heap Heap BValue) (remove-unused Heap (pointers-in Heap (pointers-in BValue ())))])
(define-metafunction vernel
  [(remove-unused (heap (Y any) ...) (X ...))
   ,(let [(xs (term (X ...)))
          (bs (term ((Y any) ...)))]
      (cons 'heap (filter (lambda (b) (member (car b) xs)) bs)))])

(define-metafunction vernel
  contains-unused : Heap BValue -> boolean
  [(contains-unused Heap BValue)
   (contains-extra Heap (pointers-in Heap (pointers-in BValue ())))])
(define-metafunction vernel
  [(contains-extra (heap (X any) ...) (Y ...))
   ,(let ([ps (term (Y ...))])
      (foldl (lambda (p acc) (or acc (not (member p ps)))) #f (term (X ...))))])

(define-metafunction vernel
  [(pointers-in FValue (X ...)) (X ...)]
  [(pointers-in Node (X ...)) (X ...)]
  [(pointers-in (Node) (X ...)) (X ...)]
  [(pointers-in (Node any_1 any_2 ...) (X ...)) (pointers-in any_1 (pointers-in (Node any_2 ...) (X ...)))]
  [(pointers-in (Heap any) (X ...)) (pointers-in Heap (pointers-in any (X ...)))]
  [(pointers-in (Env any) (X ...)) (pointers-in Env (pointers-in any (X ...)))]
  [(pointers-in (Y any) (X ...)) (pointers-in any (X ...))]
  [(pointers-in (p Y) (X ...)) (Y X ...)]
  [(pointers-in (vau (Y_1 ...) Y_2 Expr) (X ...)) (pointers-in Expr (X ...))])

(define-metafunction vernel
  [(get-call-type wrap Env Expr_1 (Expr_2 ...))
   (call-wrap Expr_1 (Env Expr_2) ...)]
  [(get-call-type wrap-ltr Env Expr_1 (Expr_2 ...))
   (call-ltr Expr_1 (Env Expr_2) ...)]
  [(get-call-type wrap-rtl Env Expr_1 (Expr_2 ...))
   (call-rtl Expr_1 (Env Expr_2) ...)]
  [(get-call-type wrap-r6rs Env Expr_1 (Expr_2 ...))
   (call-r6rs Expr_1 (thunk Env Expr_2) ...)]
  [(get-call-type wrap-lazy Env Expr_1 (Expr_2 ...))
   (build-heap call-lazy Env Expr_1 (heap) (args) Expr_2 ...)]
  [(get-call-type wrap-future Env Expr_1 (Expr_2 ...))
   (build-heap call-future Env Expr_1 (heap) (args) Expr_2 ...)])

(define-metafunction vernel
  [(delta Env Wraptype Expr) (Env (Wraptype Expr))]
  [(delta Env unwrap (Wraptype Expr)) (Env (unwrap (Wraptype Expr)))]
  [(delta Env defer any) ,(let [(s (gensym))] (term ((heap (,s (thunk Env any))) (p ,s))))]
  [(delta Env spawn any) ,(let [(s (gensym))] (term ((heap (,s (Env any))) (p ,s))))]
  [(delta Env force any) (Env (strict any))]
  [(delta Env vau (l X ...) Y Expr) (vau (X ...) Y Expr)]
  [(delta Env #t Expr_1 Expr_2) (Env Expr_1)]
  [(delta Env #f Expr_1 Expr_2) (Env Expr_2)]
  [(delta Env + number_1 number_2) ,(+ (term number_1) (term number_2))]
  [(delta Env - number_1 number_2) ,(- (term number_1) (term number_2))]
  [(delta Env * number_1 number_2) ,(* (term number_1) (term number_2))]
  [(delta Env < number_1 number_2) ,(< (term number_1) (term number_2))]
  [(delta Env > number_1 number_2) ,(> (term number_1) (term number_2))]
  [(delta Env = Value_1 Value_2) ,(eq? (term Value_1) (term Value_2))])

#|
;test Env patterns
(test-predicate list? (redex-match vernel
                                   Env (term (env))))
(test-predicate list? (redex-match vernel
                                   Env (term (env (x 1)))))
(test-predicate list? (redex-match vernel
                                   Env (term (env (x 1) (y 2)))))

;test Value patterns
(test-predicate list? (redex-match vernel
                                   Value (term 1)))
(test-predicate list? (redex-match vernel
                                   Value (term x)))
(test-predicate list? (redex-match vernel
                                   Value (term #t)))
(test-predicate list? (redex-match vernel
                                   Value (term (env))))

(test-predicate list? (redex-match vernel
                                   Value (term vau)))
(test-predicate list? (redex-match vernel
                                   Value (term (l vau (l x) y x))))

;test Expr patterns
(test-predicate list? (redex-match vernel
                                   Expr (term #t)))
(test-predicate list? (redex-match vernel
                                   Expr (term x)))
(test-predicate list? (redex-match vernel
                                   Expr (term (vau (x) y x))))
(test-predicate list? (redex-match vernel
                                   Expr (term (wrap-ltr (vau (x) y x)))))
(test-predicate list? (redex-match vernel
                                   Expr (term (unwrap (wrap-ltr (vau (x) y x))))))
(test-predicate list? (redex-match vernel
                                   Expr (term (call x))))
(test-predicate list? (redex-match vernel
                                   Expr (term (call (vau (x) y x) #t))))

;test program patterns
(test-predicate list? (redex-match vernel
                                   Prog(term ((env) #t))))
;test unboxing values
(test-->> reduce
          (term ((env) 1)) 1)
(test-->> reduce
          (term ((env) (l vau (l n) y (l n 1 2))))
          (term (vau (n) y (l n 1 2))))

;test varref
(test-->> reduce
          (term ((env (x 1)) x)) 1)
(test-->> reduce
          (term ((env (x 1)) ((env (x 1)) x))) 1)
(test-->> reduce
          (term ((env (x (vau () y y))) x)) (term (vau () y y)))

#;(apply-reduction-relation reduce
          (term ((env (x y)) z)))

;test multiple evaluation
(test-->> reduce
          (term ((env (x 1)) ((env (y x)) y))) 1)

;test conditionals
(test-->> reduce
          (term (call #t 1 2)) 1)
(test-->> reduce
          (term (call #f 1 2)) 2)

;test function lookup
(test-->> reduce
          (term ((env (x (vau () y #t))) (l x)))
          (term #t))
(test-->> reduce
          (term ((env (x (vau (x) y x))) (l x hello)))
          (term hello))
(test-->> reduce
          (term ((env (x (l a b c))) (l x hello)))
          (term ((env (x (l a b c))) (call-fn (l a b c) hello))))

;test data promotion
(test-->> reduce
          (term ((env) (l #t 1 2))) 1)

;test vau reduction
(test-->> reduce
          (term ((env) (l (l vau (l n) y (l n 1 2)) #t))) 1)

;test wrap reduction
(test-->> reduce
          (term ((env) (l (l (l wrap wrap-ltr) (l vau (l n) y (l n 1 2))) (l #t #t #f)))) 1)
(test-->> reduce
          (term ((env) (l (l (l wrap unwrap) (l (l wrap wrap-ltr) (l vau (l n) y n))) (l #t #t #f)))) (term (l #t #t #f)))
(test-->> reduce
          (term ((env) ((env) (l (l (l wrap unwrap) (l (l wrap wrap-ltr) (l vau (l n) y n))) (l #t #t #f))))) #t)

;test environment call
(test-->> reduce
          (term ((env) (l (l vau (l n) y (l y n)) (l #t 1 2)))) 1)
#;(traces reduce
          (term ((env) (l (l vau (l n) y (l y n)) (l #t 1 2)))))
|#
;test in-order evaluation
#;(test-->> reduce
            (term ((env) (l (l (l wrap wrap-ltr) (l vau (l n m) y m)) (l #t 1 2) (l #t 3 4)))) 3)
#;(traces reduce
          (term ((env) (l (l (l wrap wrap-rtl) (l vau (l n m) y m)) (l #t 1 2) (l #t 3 4)))))
#;(traces reduce
          (term ((env) (l (l (l wrap wrap) (l vau (l n m) y m)) (l #t 1 2) (l #t 3 4)))))
#;(traces reduce
          (term ((env) (l (l (l wrap wrap-r6rs) (l vau (l n m) y m)) (l #t 1 2) (l #t 3 4)))))
#;(traces reduce
          (term ((env) (l (l (l wrap wrap-future) (l vau (l n m) y m)) (l #t 1 2) (l #t 3 4)))))
#;(traces reduce
          (term ((env) (l (l (l wrap wrap) force) (l (l (l wrap wrap-lazy) (l vau (l n m) y m)) (l #t 1 2) (l #t 3 4))))))
#;(traces reduce
         (term ((env) (l (l (l wrap wrap) force) (l (l (l wrap wrap-lazy) (l vau (l m) y (l defer m))) (l #t 1 2))))))

(traces reduce
         (term ((env) (l (l (l wrap wrap) force) (l (l (l wrap wrap-lazy) (l vau (l m) y m)) (l defer (l #t 1 2)))))))


;test arithmetic
(test-->> reduce
          (term ((env) (l + 1 2))) 3)
(test-->> reduce
          (term ((env) (l (l = (l + 1 2) 0) #t #f))) #f)
