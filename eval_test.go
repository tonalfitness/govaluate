package govaluate

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEval(t *testing.T) {
	type testCase struct {
		input  string
		params map[string]interface{}
		result interface{}
	}
	testCases := [...]testCase{
		testCase{
			"floor(3.5)",
			map[string]interface{}{},
			3.0,
		},
		testCase{
			"ceil(1.1)",
			map[string]interface{}{},
			2.0,
		},
		testCase{
			"round(3.3)",
			map[string]interface{}{},
			3.0,
		},
		testCase{
			"round(4.5)",
			map[string]interface{}{},
			5.0,
		},
		testCase{
			"sqrt(4.8)",
			map[string]interface{}{},
			math.Sqrt(4.8),
		},
		testCase{
			"sin(20)",
			map[string]interface{}{},
			math.Sin(20.0),
		},
		testCase{
			"cos(20.0)",
			map[string]interface{}{},
			math.Cos(20.0),
		},
		testCase{
			"tan(20.0)",
			map[string]interface{}{},
			math.Tan(20.0),
		},
		testCase{
			"tanh(20.0)",
			map[string]interface{}{},
			math.Tanh(20.0),
		},
		testCase{
			"min(3, 4)",
			map[string]interface{}{},
			3.0,
		},
		testCase{
			"max(3, 4)",
			map[string]interface{}{},
			4.0,
		},
		testCase{
			"abs(-3)",
			map[string]interface{}{},
			3.0,
		},
		testCase{
			"log(3.0)",
			map[string]interface{}{},
			math.Log(3.0),
		},
		testCase{
			"log2(3.0)",
			map[string]interface{}{},
			math.Log2(3.0),
		},
		testCase{
			"log10(3.0)",
			map[string]interface{}{},
			math.Log10(3.0),
		},
		testCase{
			"x + y * z**2",
			map[string]interface{}{"x": -1.0, "y": 3.0, "z": 5.0},
			74.0,
		},
		testCase{
			"x > 0 ? x ** 0.5 : -x + 1",
			map[string]interface{}{"x": -6.4},
			7.4,
		},
		testCase{
			"x > 0 ? x ** 0.5 : -x + 1",
			map[string]interface{}{"x": 49.0},
			7.0,
		},
		testCase{
			"true || something",
			map[string]interface{}{},
			true,
		},
		testCase{
			"false && something",
			map[string]interface{}{},
			false,
		},
		testCase{
			"item in [1, 2, 3, 5]",
			map[string]interface{}{"item": 3.0},
			true,
		},
		testCase{
			"item in [1, 2, 3, 5]",
			map[string]interface{}{"item": 4.0},
			false,
		},
		testCase{
			"floor(a / 2) == 4",
			map[string]interface{}{"a": 9.0},
			true,
		},
		testCase{
			"a[2] + (foo ? a : b)[1+1]",
			map[string]interface{}{
				"a":   []interface{}{1.0, 2.0, 3.0},
				"b":   []interface{}{4.0, 5.0, 6.0},
				"foo": false,
			},
			9.0,
		},
		testCase{
			"a == 4",
			map[string]interface{}{"a": 4},
			true,
		},
		testCase{
			"a == 7",
			map[string]interface{}{"a": int8(7)},
			true,
		},
		testCase{
			"a == 9",
			map[string]interface{}{"a": uint(9)},
			true,
		},
		testCase{
			"a == 9",
			map[string]interface{}{"a": MustParse("(10-1)")},
			true,
		},
		testCase{
			"a == 9",
			map[string]interface{}{"a": MustParse("(b-1)"), "b": 10.0},
			true,
		},
		testCase{
			"a == 'b'",
			map[string]interface{}{"a": NewExprNodeLiteral("b", 0, 1)},
			true,
		},
		testCase{
			"a == '0x12g1'",
			map[string]interface{}{"a": TryParse("0x12g1")},
			true,
		},
	}
	for _, testCase := range testCases {
		expr, err := Parse(testCase.input)
		assert.Nil(t, err, "input=%s", testCase.input)
		val, err := expr.Eval(NewEvalParams(testCase.params))
		assert.Nil(t, err, "input=%s", testCase.input)
		assert.Equal(t, testCase.result, val, "input=%s", testCase.input)
	}
}

func TestEvalError(t *testing.T) {
	type testCase struct {
		input  string
		params map[string]interface{}
		err    string
	}
	testCases := [...]testCase{
		testCase{
			"x + y * (z**2 > 0)",
			map[string]interface{}{"x": 1.0, "y": 2.0, "z": 3.0},
			"rhs of + / rhs of * is not numeric: true [pos=8; len=10]",
		},
		testCase{
			"x ? 1 : 0",
			map[string]interface{}{"x": 1.0},
			"ternary condition is not boolean: 1 [pos=0; len=1]",
		},
		testCase{
			"[1, arr[0], 3]",
			map[string]interface{}{"arr": 1.0},
			"array item #2 / indexer receiver is not array: 1 [pos=4; len=3]",
		},
		testCase{
			"2**floor(x, y)",
			map[string]interface{}{},
			"rhs of ** / wrong number of arguments: 2, expected: 1 [op=floor; pos=3; len=11]",
		},
		testCase{
			"[1, 2, 3][3] * 2",
			map[string]interface{}{},
			"lhs of * / index out of bounds: 3, len: 3 [op=[]; pos=0; len=12]",
		},
	}

	for _, testCase := range testCases {
		expr, err := Parse(testCase.input)
		assert.Nil(t, err, "input=%s", testCase.input)
		_, err = expr.Eval(NewEvalParams(testCase.params))
		assert.EqualError(t, err, testCase.err, "input=%s", testCase.input)
	}
}
