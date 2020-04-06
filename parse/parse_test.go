package parse

import (
	"fmt"
	"github.com/mchaynes/funnel/lex"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct{
		Input []lex.Item
		Output *BooleanNode
	}{
		{
			Input: []lex.Item{
				{lex.OpenPrecedenceType, "("},
				{lex.FieldType, "Key"},
				{lex.ComparatorType, "<="},
				{lex.LiteralType, `"12"`},
				{lex.BooleanType, "&&"},
				{lex.LiteralType, `"15"`},
				{lex.ComparatorType, "<="},
				{lex.FieldType, "Key2"},
				{lex.ClosePrecedenceType, ")"},
			},
			Output: &BooleanNode{
				Operation: "&&",
				Children:  []Criteria{
					&BooleanNode{
						Operation: "&&",
						Children: []Criteria{
							&FieldNode{
								Field: "Key",
								Comp:  "<=",
								Value: "12",
							},
							&FieldNode{
								Field: "Key2",
								Comp:  "<=",
								Value: "15",
							},
						},
					},
				},
			},
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ch := Fill(c.Input)
			criteria, err := Parse(ch)
			assert.NoError(t, err)
			bn := criteria.(*BooleanNode)
			assert.Equal(t, *c.Output, *bn)
		})
	}
}


func Fill(items []lex.Item) chan lex.Item {
	c := make(chan lex.Item, len(items))
	for _, i := range items {
		c <- i
	}
	close(c)
	return c
}
