package needle

import (
	"errors"
	"fmt"
	"needle/internal/needle/evaluator"
	"needle/internal/needle/lexer"
	"needle/internal/needle/parser"
)

type Needle struct {
	ev   *evaluator.Evaluator
	glob *evaluator.Env
}

func New() *Needle {
	glob := evaluator.NewEnv(nil)
	return &Needle{
		ev:   evaluator.New(),
		glob: glob,
	}
}

func LoadBuiltin(n *Needle) {
	evaluator.LoadBuiltins(n.ev)
}

func (n *Needle) LoadFunction(
	name string,
	f evaluator.NativeFunction,
	arity int,
) {
	nf := &evaluator.Function{
		FType:  evaluator.F_NATIVE,
		Native: coverNative(f, arity),
	}
	n.glob.Declare(name, nf)
}

func (n *Needle) RunString(source string) error {
	script, errs := createAST([]byte(source))
	for _, err := range errs {
		fmt.Println(err)
	}
	if errs != nil {
		return errors.New("compile error")
	}
	return n.ev.Run(script)
}

func createAST(source []byte) (*parser.Script, []error) {
	lx := lexer.New([]rune(string(source)))
	return parser.New(lx).Parse()
}

func coverNative(f evaluator.NativeFunction, a int) evaluator.NativeFunction {
	return func(
		e *evaluator.Evaluator,
		this evaluator.Value,
		args ...evaluator.Value,
	) evaluator.Value {
		err := evaluator.CheckArgsLength(a, args)
		if err != nil {
			e.ThrowException("%s", err.Error())
		}
		return f(e, this, args...)
	}
}
