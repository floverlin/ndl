package needle

import (
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
		ev:   evaluator.New(glob),
		glob: glob,
	}
}

func (n *Needle) LoadFunction(
	name string,
	f evaluator.NativeFunction,
	arity uint,
) {
	nf := &evaluator.Function{
		FType:  evaluator.F_NATIVE,
		Native: f,
	}
	n.glob.Declare(name, nf, true)
}

func (n *Needle) RunString(source string) error {
	script, err := createAST([]byte(source))
	if err != nil {
		return err
	}
	return n.ev.Run(script)
}

func createAST(source []byte) (*parser.Script, error) {
	lx := lexer.New([]rune(string(source)))
	return parser.New(lx).Parse()
}
