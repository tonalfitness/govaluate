package govaluate

import (
	"fmt"
	"math"
)

func BuiltinOperators() map[string]Operator {
	return map[string]Operator{
		"==": BinaryOp(func(a, b interface{}) interface{} {
			return a == b
		}),
		"!=": BinaryOp(func(a, b interface{}) interface{} {
			return a != b
		}),
		">": BinaryNumericOp(func(a, b float64) interface{} {
			return a > b
		}),
		"<": BinaryNumericOp(func(a, b float64) interface{} {
			return a < b
		}),
		">=": BinaryNumericOp(func(a, b float64) interface{} {
			return a >= b
		}),
		"<=": BinaryNumericOp(func(a, b float64) interface{} {
			return a <= b
		}),
		"&&": BooleanFold(true, func(a, b bool) (bool, bool) {
			return a && b, !a || !b
		}),
		"||": BooleanFold(false, func(a, b bool) (bool, bool) {
			return a || b, a || b
		}),
		"+": BinaryNumericOp(func(a, b float64) interface{} {
			return a + b
		}),
		"-": func(args []LazyArg) (interface{}, error) {
			if len(args) == 1 {
				right, err := args[0].GetNumeric()
				if err != nil {
					return nil, err
				}
				return -right, nil
			}
			if len(args) == 2 {
				left, err := args[0].GetNumeric()
				if err != nil {
					return nil, err
				}
				right, err := args[1].GetNumeric()
				if err != nil {
					return nil, err
				}
				return left - right, nil
			}
			return nil, fmt.Errorf("wrong number of arguments: %d", len(args))
		},
		"&": BinaryNumericOp(func(a, b float64) interface{} {
			return float64(int(a) & int(b))
		}),
		"|": BinaryNumericOp(func(a, b float64) interface{} {
			return float64(int(a) | int(b))
		}),
		"^": BinaryNumericOp(func(a, b float64) interface{} {
			return float64(int(a) ^ int(b))
		}),
		"<<": BinaryNumericOp(func(a, b float64) interface{} {
			return float64(int(a) << uint(b))
		}),
		">>": BinaryNumericOp(func(a, b float64) interface{} {
			return float64(int(a) >> uint(b))
		}),
		"*": BinaryNumericOp(func(a, b float64) interface{} {
			return a * b
		}),
		"/": BinaryNumericOp(func(a, b float64) interface{} {
			return a / b
		}),
		"%": BinaryNumericOp(func(a, b float64) interface{} {
			return float64(int(a) % int(b))
		}),
		"**": BinaryNumericOp(func(a, b float64) interface{} {
			return math.Pow(a, b)
		}),
		"~": UnaryNumericOp(func(a float64) interface{} {
			return float64(^int(a))
		}),
		"!": UnaryBooleanOp(func(a bool) interface{} {
			return !a
		}),
		"if": func(args []LazyArg) (interface{}, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("wrong number of arguments: %d", len(args))
			}
			condition, err := args[0].GetBoolean()
			if err != nil {
				return nil, err
			}
			if condition {
				return args[1]()
			}
			return args[2]()
		},
		"??": func(args []LazyArg) (interface{}, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("wrong number of arguments: %d", len(args))
			}
			left, err := args[0]()
			if err != nil || left != nil {
				return left, err
			}
			return args[1]()
		},
	}
}

func BinaryOp(fn func(interface{}, interface{}) interface{}) Operator {
	return func(args []LazyArg) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("wrong number of arguments: %d", len(args))
		}
		left, err := args[0]()
		if err != nil {
			return nil, err
		}
		right, err := args[1]()
		if err != nil {
			return nil, err
		}
		return fn(left, right), nil
	}
}

func BinaryNumericOp(fn func(float64, float64) interface{}) Operator {
	return func(args []LazyArg) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("wrong number of arguments: %d", len(args))
		}
		left, err := args[0].GetNumeric()
		if err != nil {
			return nil, err
		}
		right, err := args[1].GetNumeric()
		if err != nil {
			return nil, err
		}
		return fn(left, right), nil
	}
}

func BooleanFold(initial bool, fn func(bool, bool) (bool, bool)) Operator {
	return func(args []LazyArg) (interface{}, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("empty args")
		}
		acc := initial
		for _, arg := range args {
			val, err := arg.GetBoolean()
			if err != nil {
				return nil, err
			}
			newAcc, stop := fn(acc, val)
			if stop {
				return newAcc, nil
			}
			acc = newAcc
		}
		return acc, nil
	}
}

func UnaryNumericOp(fn func(float64) interface{}) Operator {
	return func(args []LazyArg) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("wrong number of arguments: %d", len(args))
		}
		right, err := args[0].GetNumeric()
		if err != nil {
			return nil, err
		}
		return fn(right), nil
	}
}

func UnaryBooleanOp(fn func(bool) interface{}) Operator {
	return func(args []LazyArg) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("wrong number of arguments: %d", len(args))
		}
		right, err := args[0].GetBoolean()
		if err != nil {
			return nil, err
		}
		return fn(right), nil
	}
}
