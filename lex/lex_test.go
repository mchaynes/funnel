package lex_test

import (
	"github.com/mchaynes/funnel/lex"
	"github.com/stretchr/testify/assert"
	"testing"
)



func TestLexer_Run(t *testing.T) {
	cases := []struct{
		Input string
		Expected []lex.Item
	}{
		{
			Input: `(Key = "Value")`,
			Expected: []lex.Item{
				{lex.OpenPrecedenceType, "("},
				{lex.FieldType, "Key"},
				{lex.ComparatorType, "="},
				{lex.LiteralType, `"Value"`},
				{lex.ClosePrecedenceType, ")"},
			},
		},
		{
			Input: `( "Value" <= Key )`,
			Expected: []lex.Item{
				{lex.OpenPrecedenceType, "("},
				{lex.LiteralType, `"Value"`},
				{lex.ComparatorType, "<="},
				{lex.FieldType, "Key"},
				{lex.ClosePrecedenceType, ")"},
			},
		},
		{
			Input: `"12" > Data.Value && ( Data.Value < "15" || "19" > Data.Value )`,
			Expected: []lex.Item{
				{lex.LiteralType, `"12"`},
				{lex.ComparatorType, `>`},
				{lex.FieldType, "Data.Value"},
				{lex.BooleanType, "&&"},
				{lex.OpenPrecedenceType, "("},
				{lex.FieldType, "Data.Value"},
				{lex.ComparatorType, "<"},
				{lex.LiteralType, `"15"`},
				{lex.BooleanType, "||"},
				{lex.LiteralType, `"19"`},
				{lex.ComparatorType, ">"},
				{lex.FieldType, "Data.Value"},
				{lex.ClosePrecedenceType, ")"},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Input, checkItems(c.Input, buildDict(c.Expected), c.Expected))
	}
}

func checkItems(input string, allowed []string, ex []lex.Item) func(t *testing.T) {
	return func(t *testing.T) {
		l := lex.NewLexer(input, allowed)
		i := 0
		for a := l.NextItem(); a != nil; a = l.NextItem() {
			if i < len(ex) {
				e := ex[i]
				assert.Equal(t, e.Value, a.Value, "values should have matched")
				assert.Equal(t, e.Type, a.Type, "type should have matched")
				i++
			} else {
				t.Errorf("extra unexpected item returned %v", a)
			}
		}
		assert.Equal(t, len(ex), i, "l.Items() should have returned the proper number of results")
	}
}

func TestLexer_NextItem(t *testing.T) {
	cases := []struct{
		input string
		output []lex.Item
	}{
		{
			input: `Hello < "123"`,
			output: []lex.Item{
				{Type: lex.FieldType, Value: "Hello"},
				{Type: lex.ComparatorType, Value: "<"},
				{Type: lex.LiteralType, Value: `"123"`},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			lexer := lex.NewLexer(c.input, buildDict(c.output))
			for i := 0; i < len(c.output); i++ {
				item := c.output[i]
				actual := lexer.NextItem()
				assert.Equal(t, item.Value, actual.Value)
				assert.Equal(t, item.Type, actual.Type)
			}
		})
	}
}

func buildDict(items []lex.Item) []string {
	m := make([]string, 0)
	for _, i := range items {
		if i.Type == lex.FieldType {
			m = append(m, i.Value)
		}
	}
	return m
}
