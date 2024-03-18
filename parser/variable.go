package parser

import "fmt"

// Variable represents variable name and type
// if variable is a array  or map type, Index or IndexName or IndexnVariableName must be set
type Variable struct {
	Name               string
	VariableType       PrimaryType
	Index              int
	IndexnName         string
	IndexnVariableName string
	Value              any
}

func NewVariable() *Variable {
	return &Variable{
		VariableType: UNKNOWN,
		Index:        -1,
	}
}

func (d *Variable) String() string {
	return fmt.Sprintf("name: %s, type:%s, index:%d, index_name:%s, index_variable_name:%s, value:%v", d.Name, d.VariableType.String(), d.Index, d.IndexnName, d.IndexnVariableName, d.Value)
}

func (d *Variable) Equal(other *Variable) bool {
	return d.Name == other.Name && d.VariableType == other.VariableType &&
		d.Index == other.Index && d.IndexnName == other.IndexnName && d.Value == other.Value &&
		d.IndexnVariableName == other.IndexnVariableName
}
