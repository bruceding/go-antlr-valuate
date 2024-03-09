package parser

import (
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
)

type VariableScanListener struct {
	BaseGovaluateListener
	currentVariable *Variable
	node2Variables  map[antlr.ParserRuleContext]*Variable
}

func NewVariableScanListener() *VariableScanListener {

	return &VariableScanListener{
		node2Variables: make(map[antlr.ParserRuleContext]*Variable),
	}
}

func (s *VariableScanListener) EnterExpression(ctx *ExpressionContext) {
	if ctx.Primary() != nil {
		s.currentVariable = NewVariable()
	}
}
func (s *VariableScanListener) ExitExpression(ctx *ExpressionContext) {
	if ctx.Primary() != nil && s.currentVariable != nil {
		s.node2Variables[ctx] = s.currentVariable
	}
	if len(ctx.AllExpression()) == 2 {
		switch ctx.bop.GetText() {
		case "[": // maybe a slice or a map
			leftVariable := s.node2Variables[ctx.Expression(0)]
			rightVariable := s.node2Variables[ctx.Expression(1)]
			if leftVariable != nil && rightVariable != nil {
				leftVariable.VariableType = ARRAYORMAP
				if rightVariable.VariableType == FLOAT {
					leftVariable.Index = int(rightVariable.Value.(float64))
				} else if rightVariable.VariableType == STRING {
					leftVariable.IndexnName = rightVariable.Value.(string)
				}

				s.node2Variables[ctx] = leftVariable
			}

		}
	}
}
func (s *VariableScanListener) ExitIdentifier(ctx *IdentifierContext) {
	str := ctx.GetText()
	str = strings.Trim(str, "${}")
	if str != "" {
		if s.currentVariable != nil {
			s.currentVariable.Name = str
		}
	}

}

func (s *VariableScanListener) ExitString(ctx *StringContext) {
	str := ctx.GetText()
	str = strings.Trim(str, "'")
	if str != "" {
		if s.currentVariable != nil {
			s.currentVariable.Value = str
			s.currentVariable.VariableType = STRING
		}
	}
}

func (s *VariableScanListener) ExitFloat(ctx *FloatContext) {
	str := ctx.GetText()
	if f, err := strconv.ParseFloat(str, 64); err == nil {
		if s.currentVariable != nil {
			s.currentVariable.VariableType = FLOAT
			s.currentVariable.Value = f
		}
	}
}
