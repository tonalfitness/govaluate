package govaluate

import "fmt"

type EvalParams struct {
	Variables map[string]interface{}
	Operators map[string]Operator
}

func (expr ExprNode) Eval(params EvalParams) (interface{}, error) {
	switch expr.Type {
	case NodeTypeLiteral:
		return expr.Value, nil
	case NodeTypeVariable:
		value, ok := params.Variables[expr.Name]
		if !ok {
			return nil, fmt.Errorf("variable undefined: %v [pos=%d; len=%d]", expr.Name, expr.SourcePos, expr.SourceLen)
		}

		val, stringType := value.(string)
		if !stringType {
			return value, nil
		}

		// Try to parse string in case var is an additional expression
		node, err := Parse(val)
		if err != nil {
			return value, nil
		}
		for _, v := range node.Vars() {
			if v == expr.Name {
				return nil, fmt.Errorf("variable can not refer to itself: %v [pos=%d; len=%d]", expr.Name, expr.SourcePos, expr.SourceLen)
			}
		}
		return node.Eval(params)
	case NodeTypeOperator:
		operator, ok := params.Operators[expr.Name]
		if !ok {
			return nil, fmt.Errorf("operator undefined: %v [pos=%d; len=%d]", expr.Name, expr.SourcePos, expr.SourceLen)
		}
		return operator(EvalContext{params: params, expr: expr})
	}
	return nil, fmt.Errorf("bad expr type: %v", expr)
}

var builtinOperators = BuiltinOperators()

func NewEvalParams(variables map[string]interface{}) EvalParams {
	return EvalParams{
		Variables: variables,
		Operators: builtinOperators,
	}
}
