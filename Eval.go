package govaluate

import "fmt"

type LazyArg func() (interface{}, error)
type Operator func([]LazyArg) (interface{}, error)

type EvalParams struct {
	Variables map[string]interface{}
	Operators map[string]Operator
}

func (node ExprNode) Eval(params EvalParams) (interface{}, error) {
	switch node.Type {
	case NodeTypeLiteral:
		return node.Value, nil
	case NodeTypeVariable:
		value, ok := params.Variables[node.Name]
		if !ok {
			return nil, fmt.Errorf("variable undefined: %v", node.Name)
		}
		return value, nil
	case NodeTypeOperator:
		operator, ok := params.Operators[node.Name]
		if !ok {
			return nil, fmt.Errorf("operator undefined: %v", node.Name)
		}

		lazyArgs := make([]LazyArg, len(node.Args))
		for idx, arg := range node.Args {
			lazyArgs[idx] = NewLazyArg(arg, params)
		}

		return operator(lazyArgs)
	}
	return nil, fmt.Errorf("bad node type: %v", node)
}

func NewEvalParams(variables map[string]interface{}) EvalParams {
	return EvalParams{
		Variables: variables,
		Operators: BuiltinOperators(),
	}
}

func NewLazyArg(node ExprNode, params EvalParams) LazyArg {
	return func() (interface{}, error) {
		return node.Eval(params)
	}
}

func (arg LazyArg) GetNumeric() (float64, error) {
	val, err := arg()
	if err != nil {
		return 0.0, err
	}
	if numericVal, ok := val.(float64); ok {
		return numericVal, nil
	}
	return 0.0, fmt.Errorf("expected number, but got: %v", val)
}

func (arg LazyArg) GetBoolean() (bool, error) {
	val, err := arg()
	if err != nil {
		return false, err
	}
	if booleanVal, ok := val.(bool); ok {
		return booleanVal, nil
	}
	return false, fmt.Errorf("expected boolean, but got: %v", val)
}
