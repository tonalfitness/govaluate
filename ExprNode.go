package govaluate

import (
	"fmt"
)

// ExprNode is a structured representation of an expression.
// There are three types of nodes: literal, variable and operator. The latter
// can have child nodes. They form a tree, where each node is an expression itself.
type ExprNode struct {
	Type  ExprNodeType
	Name  string
	Value interface{}
	Args  []ExprNode
}

// ExprNodeType is a type of ExprNode.
type ExprNodeType int

const (
	// NodeTypeLiteral is just a constant literal, e.g. a boolean, a number, or a string.
	// ExprNode.Value contains the actual value.
	NodeTypeLiteral ExprNodeType = iota

	// NodeTypeVariable is a variable.
	// ExprNode.Name contains the name of the variable.
	NodeTypeVariable

	// NodeTypeOperator is an operation over the arguments, the other nodes.
	// It can be a function call, a binary operator (+, -, *, etc) with two arguments,
	// unary (!, -, ~), ternary (?:), etc. There are no restrictions on operator name,
	// it just needs to be defined at the evaluation phase.
	// ExprNode.Name is the name of the operation.
	// ExprNode.Args are the arguments.
	NodeTypeOperator
)

// NewExprNodeLiteral constructs a literal node.
func NewExprNodeLiteral(value interface{}) ExprNode {
	return ExprNode{
		Type:  NodeTypeLiteral,
		Value: value,
	}
}

// NewExprNodeVariable constructs a variable node.
func NewExprNodeVariable(name string) ExprNode {
	return ExprNode{
		Type: NodeTypeVariable,
		Name: name,
	}
}

// NewExprNodeOperator constructs an operator node.
func NewExprNodeOperator(name string, args ...ExprNode) ExprNode {
	return ExprNode{
		Type: NodeTypeOperator,
		Name: name,
		Args: args,
	}
}

// IsOperator returns true if this node is an operator with matching name.
func (node ExprNode) IsOperator(name string) bool {
	return node.Type == NodeTypeOperator && node.Name == name
}

// IsLiteral returns true if this node is a literal with matching value.
func (node ExprNode) IsLiteral(value interface{}) bool {
	return node.Type == NodeTypeLiteral && node.Value == value
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
		// binary operator or ternary if
		left, err := stage.leftStage.ToExprNode()
		if err != nil {
			return ExprNode{}, err
		}
		right, err := stage.rightStage.ToExprNode()
		if err != nil {
			return ExprNode{}, err
		}
		if stage.symbol == TERNARY_FALSE {
			// ternary if
			if !left.IsOperator("?") {
				return ExprNode{}, fmt.Errorf("unexpected ternary: %v", left)
			}
			return NewExprNodeOperator("if", left.Args[0], left.Args[1], right), nil
		}
		// binary operator
		opName := stage.symbol.String()
		if stage.symbol == EQ {
			opName = "==" // for some reason EQ.String() is '='
		}
		return NewExprNodeOperator(opName, left, right), nil

	case NEGATE, INVERT, BITWISE_NOT:
		// unary operator
		right, err := stage.rightStage.ToExprNode()
		if err != nil {
			return ExprNode{}, err
		}
		opName := stage.symbol.String()
		return NewExprNodeOperator(opName, right), nil
	}

	return ExprNode{}, fmt.Errorf("unknown symbol: %v", stage.symbol)
}

type paramCaptor struct {
	lastParameterName string
}

func (p *paramCaptor) Get(key string) (interface{}, error) {
	p.lastParameterName = key
	return 0.0, nil
}
