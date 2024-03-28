package valuate

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
)

type EvaluableStatementErrorListener struct {
	antlr.DefaultErrorListener
	err error
}

func NewEvaluableStatementErrorListener() *EvaluableStatementErrorListener {
	return &EvaluableStatementErrorListener{}
}

func (d *EvaluableStatementErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	err := fmt.Errorf("parse error at line %d:%d %s", line, column, msg)
	d.err = err
}

func (d *EvaluableStatementErrorListener) Error() error {
	return d.err
}
