package funnel

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestBoolean_ToSql(t *testing.T) {
	var (
		helloF = field{field: "Hello", comp:"<", val: "123"}
		byeF = field{field: "Goodbye", comp:"=", val: "World"}
		alohaF = field{field:"Aloha", comp:">=", val:"Hello"}
		nestedHiBye = boolean{operation:Or, children:[]sqler{&helloF, &byeF}}
	)
	var cases = []struct{
		b boolean
		output string
		args []interface{}
	}{
		{
			b: boolean{operation:And, children: []sqler{&helloF}},
			output: "( Hello < ? )",
			args: []interface{}{helloF.val},
		},
		{
			b: boolean{operation:And, children: []sqler{&byeF, &helloF}},
			output: "( Goodbye = ? AND Hello < ? )",
			args: []interface{}{byeF.val, helloF.val},
		},
		{
			b: boolean{operation:Or, children: []sqler{&alohaF, &byeF, &helloF}},
			output: "( Aloha >= ? OR Goodbye = ? OR Hello < ? )",
			args: []interface{}{alohaF.val, byeF.val, helloF.val},
		},
		{
			b: boolean{operation:And, children:[]sqler{&nestedHiBye, &alohaF}},
			output: "( ( Hello < ? OR Goodbye = ? ) AND Aloha >= ? )",
			args: []interface{}{helloF.val, byeF.val, alohaF.val},
		},
	}

	for _, c := range cases {
		t.Run(c.output, func(t *testing.T) {
			actual, args := c.b.ToSql()
			assert.Equal(t, c.output, actual)
			assert.Equal(t, c.args, args)
		})
	}
}


func TestToSql(t *testing.T) {
	cases := []struct{
		input string
		output string
		args []interface{}
	}{
		{
			input: `Hello <= "123" && ( Goodbye = "Farewell" || Aloha > "Hello" )`,
			output: "( Hello <= ? AND ( Goodbye = ? OR Aloha > ? ) )",
			args: []interface{}{"123", "Farewell", "Hello"},
		},
	}
	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			sql, args, err := ToSql(c.input, strings.Split(c.input, " "))
			assert.NoError(t, err)
			assert.Equal(t, c.output, sql)
			assert.Equal(t, c.args, args)
		})
	}
}

func TestReplaceDollar(t *testing.T) {
	cases := map[string]string{
		"Hello = ?": "Hello = $1",
		"Hello = ? AND Goodbye = ?": "Hello = $1 AND Goodbye = $2",
		"A = ? AND B = ? AND C = ?": "A = $1 AND B = $2 AND C = $3",
	}
	for in, expected := range cases {
		t.Run(in, func(t *testing.T) {
			out, err := Replace(in, Dollar)
			assert.NoError(t, err)
			assert.Equal(t, expected, out)
		})
	}
}
