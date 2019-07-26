package govaluate

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

type ExprNodePrinter struct {
	nodeHandler func(ExprNode, *ExprNodePrinter) error
	output      strings.Builder
	err         error
}

type PrintConfig struct {
	FormatBoolLiteral   func(bool) string
	FormatNumberLiteral func(float64) string
	FormatStringLiteral func(string) string
	FormatVariable      func(string) string
	OperatorMap         map[string]string
	OperatorMapper      func(name string, arity int) string
	InfixOperators      map[string]bool
	PrecedenceFn        func(name string, arity int) int
	Operators           map[string]func(args []ExprNode, output *ExprNodePrinter) error
}

func (b *ExprNodePrinter) AppendString(token string) {
	if b.err == nil {
		b.output.WriteString(token)
	}
}

func (b *ExprNodePrinter) AppendNode(node ExprNode) {
	if b.err == nil {
		err := b.nodeHandler(node, b)
		if b.err == nil {
			b.err = err
		}
	}
}

func (expr EvaluableExpression) Print(config PrintConfig) (string, error) {
	return expr.PrintWithHandler(defaultNodeHandler(config))
}

func (expr EvaluableExpression) PrintWithHandler(nodeHandler func(ExprNode, *ExprNodePrinter) error) (string, error) {
	node, err := expr.evaluationStages.ToExprNode()
	if err != nil {
		return "", err
	}
	return node.PrintWithHandler(nodeHandler)
}

func (node ExprNode) Print(config PrintConfig) (string, error) {
	return node.PrintWithHandler(defaultNodeHandler(config))
}

func (node ExprNode) PrintWithHandler(nodeHandler func(ExprNode, *ExprNodePrinter) error) (string, error) {
	builder := &ExprNodePrinter{nodeHandler: nodeHandler}
	builder.AppendNode(node)
	return builder.output.String(), builder.err
}

func defaultNodeHandler(config PrintConfig) func(ExprNode, *ExprNodePrinter) error {
	return func(node ExprNode, output *ExprNodePrinter) error {
		switch node.Type {
		case NodeTypeLiteral:
			return literal(node.Value, output, &config)
		case NodeTypeVariable:
			return variable(node.Name, output, &config)
		case NodeTypeOperator:
			return operator(node.Name, node.Args, output, &config)
		}
		return fmt.Errorf("unexpected node: %v", node)
	}
}

func literal(value interface{}, output *ExprNodePrinter, config *PrintConfig) error {
	var literal string
	switch value.(type) {
	case bool:
		literal = boolLiteral(value.(bool), config)
	case float64:
		literal = numberLiteral(value.(float64), config)
	case string:
		literal = stringLiteral(value.(string), config)
	default:
		return fmt.Errorf("unsupported literal type: %v", value)
	}
	output.AppendString(literal)
	return nil
}

func boolLiteral(value bool, config *PrintConfig) string {
	if config.FormatBoolLiteral != nil {
		return config.FormatBoolLiteral(value)
	}
	if value {
		return "true"
	}
	return "false"
}

func numberLiteral(value float64, config *PrintConfig) string {
	if config.FormatNumberLiteral != nil {
		return config.FormatNumberLiteral(value)
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func stringLiteral(value string, config *PrintConfig) string {
	if config.FormatStringLiteral != nil {
		return config.FormatStringLiteral(value)
	}
	escapedValue := strings.NewReplacer(
		"\\", "\\\\",
		"\"", "\\\"",
		"\r", "\\r",
		"\n", "\\n",
	).Replace(value)
	return "\"" + escapedValue + "\""
}

func variable(name string, output *ExprNodePrinter, config *PrintConfig) error {
	variable := name
	if config.FormatVariable != nil {
		variable = config.FormatVariable(name)
	}
	output.AppendString(variable)
	return nil
}

func operator(name string, args []ExprNode, output *ExprNodePrinter, config *PrintConfig) error {
	arity := len(args)
	mappedName := config.mappedName(name, arity)

	if fn, ok := config.Operators[mappedName]; ok {
		return fn(args, output)
	}

	// binary operator: x + y
	infix := config.isInfix(name, arity)
	if infix {
		selfPrecedence := config.precedence(name, arity)
		leftPrecedence := config.precedenceForNode(args[0])
		rightPrecedence := config.precedenceForNode(args[1])
		if leftPrecedence < selfPrecedence {
			output.AppendString("(")
		}
		output.AppendNode(args[0])
		if leftPrecedence < selfPrecedence {
			output.AppendString(")")
		}
		output.AppendString(" ")
		output.AppendString(mappedName)
		output.AppendString(" ")
		if rightPrecedence <= selfPrecedence {
			output.AppendString("(")
		}
		output.AppendNode(args[1])
		if rightPrecedence <= selfPrecedence {
			output.AppendString(")")
		}
		return nil
	}

	// prefix operator: !x
	prefix := arity == 1 && isSpecial(mappedName)
	if prefix {
		selfPrecedence := config.precedence(name, arity)
		rightPrecedence := config.precedenceForNode(args[0])
		output.AppendString(mappedName)
		if rightPrecedence < selfPrecedence {
			output.AppendString("(")
		}
		output.AppendNode(args[0])
		if rightPrecedence < selfPrecedence {
			output.AppendString(")")
		}
		return nil
	}

	// ternary if: x ? y : z
	if mappedName == "if" && arity == 3 {
		selfPrecedence := config.precedence(name, arity)
		conditionPrecedence := config.precedenceForNode(args[0])
		thenPrecedence := config.precedenceForNode(args[1])
		elsePrecedence := config.precedenceForNode(args[2])
		if conditionPrecedence <= selfPrecedence {
			output.AppendString("(")
		}
		output.AppendNode(args[0])
		if conditionPrecedence <= selfPrecedence {
			output.AppendString(")")
		}
		output.AppendString(" ? ")
		if thenPrecedence <= selfPrecedence {
			output.AppendString("(")
		}
		output.AppendNode(args[1])
		if thenPrecedence <= selfPrecedence {
			output.AppendString(")")
		}
		output.AppendString(" : ")
		if elsePrecedence < selfPrecedence {
			output.AppendString("(")
		}
		output.AppendNode(args[2])
		if elsePrecedence < selfPrecedence {
			output.AppendString(")")
		}
		return nil
	}

	// function call: fn(a, b, c)
	output.AppendString(mappedName)
	output.AppendString("(")
	for idx, arg := range args {
		if idx > 0 {
			output.AppendString(", ")
		}
		output.AppendNode(arg)
	}
	output.AppendString(")")
	return nil
}

func isSpecial(name string) bool {
	for _, r := range []rune(name) {
		if unicode.IsLetter(r) {
			return false
		}
	}
	return len(name) > 0
}

func (config *PrintConfig) mappedName(operator string, arity int) string {
	if mappedName, ok := config.OperatorMap[operator]; ok {
		return mappedName
	}
	if config.OperatorMapper != nil {
		if mappedName := config.OperatorMapper(operator, arity); mappedName != "" {
			return mappedName
		}
	}
	return operator
}

func (config *PrintConfig) isInfix(operator string, arity int) bool {
	if arity != 2 {
		return false
	}
	mappedName := config.mappedName(operator, arity)
	if infix, found := config.InfixOperators[mappedName]; found {
		return infix
	}
	return isSpecial(mappedName) || mappedName == "in"
}

func (config *PrintConfig) precedenceForNode(node ExprNode) int {
	if node.Type == NodeTypeOperator {
		return config.precedence(node.Name, len(node.Args))
	}
	// variable and literal have max precedence
	return math.MaxInt32
}

func (config *PrintConfig) precedence(operator string, arity int) int {
	if config.PrecedenceFn != nil {
		mappedName := config.mappedName(operator, arity)
		return config.PrecedenceFn(mappedName, arity)
	}
	switch operator {
	case "if":
		return 0
	case "??":
		return 1
	case "||":
		return 2
	case "&&":
		return 3
	case "==", "!=", ">", "<", ">=", "<=", "=~", "!~", "in":
		return 4
	case "&", "|", "^", "<<", ">>":
		return 5
	case "+":
		return 6
	case "-":
		if arity == 1 {
			// unary minus
			return 8
		}
		return 6
	case "*", "/", "%":
		return 7
	case "!", "~":
		return 8
	case "**":
		return 9
	}
	return 10
}
