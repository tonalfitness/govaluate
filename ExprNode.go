package govaluate

import (
	"fmt"
)

type ExprNodeType int

const (
	NodeTypeLiteral ExprNodeType = iota
	NodeTypeVariable
	NodeTypeOperator
)

type ExprNode struct {
	Type  ExprNodeType
	Name  string
	Value interface{}
	Args  []ExprNode
}

func NewExprNodeLiteral(value interface{}) ExprNode {
	return ExprNode{
		Type:  NodeTypeLiteral,
		Value: value,
	}
}

func NewExprNodeVariable(name string) ExprNode {
	return ExprNode{
		Type: NodeTypeVariable,
		Name: name,
	}
}

func NewExprNodeOperator(name string, args ...ExprNode) ExprNode {
	return ExprNode{
		Type: NodeTypeOperator,
		Name: name,
		Args: args,
	}
}

func (node ExprNode) IsOperator(name string) bool {
	return node.Type == NodeTypeOperator && node.Name == name
}

func (stage *evaluationStage) ToExprNode() (ExprNode, error) {
	if stage == nil {
		return ExprNode{}, fmt.Errorf("stage == nil")
	}

	switch stage.symbol {
	case VALUE:
		// get variable name: invoke the operator and look what parameter it has queried
		p := paramCaptor{}
		if _, err := stage.operator(nil, nil, &p); err != nil {
			return ExprNode{}, err
		}
		variableName := p.lastParameterName
		return NewExprNodeVariable(variableName), nil

	case LITERAL:
		// get literal value
		value, err := stage.operator(nil, nil, nil)
		if err != nil {
			return ExprNode{}, err
		}
		return NewExprNodeLiteral(value), nil

	case NOOP:
		// parentheses
		return stage.rightStage.ToExprNode()

	case EQ, NEQ, GT, LT, GTE, LTE, REQ, NREQ, IN, AND, OR,
		PLUS, MINUS, BITWISE_AND, BITWISE_OR, BITWISE_XOR,
		BITWISE_LSHIFT, BITWISE_RSHIFT, MULTIPLY, DIVIDE,
		MODULUS, EXPONENT, TERNARY_TRUE, TERNARY_FALSE, COALESCE:
		left, err := stage.leftStage.ToExprNode()
		if err != nil {
			return ExprNode{}, err
		}
		right, err := stage.rightStage.ToExprNode()
		if err != nil {
			return ExprNode{}, err
		}
		if stage.symbol == TERNARY_FALSE {
			if !left.IsOperator("?") {
				return ExprNode{}, fmt.Errorf("unexpected ternary: %v", left)
			}
			return NewExprNodeOperator("if", left.Args[0], left.Args[1], right), nil
		}
		opName := stage.symbol.String()
		if stage.symbol == EQ {
			opName = "==" // for some reason EQ.String() is '='
		}
		return NewExprNodeOperator(opName, left, right), nil

	case NEGATE, INVERT, BITWISE_NOT:
		right, err := stage.rightStage.ToExprNode()
		if err != nil {
			return ExprNode{}, err
		}
		opName := stage.symbol.String()
		return NewExprNodeOperator(opName, right), nil
	}

	return ExprNode{}, fmt.Errorf("unknown symbol: %v", stage.symbol)
}
