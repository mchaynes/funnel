package parse

import (
	"fmt"
	"github.com/mchaynes/funnel/lex"
	"strings"
)

type Criteria interface {
	Matches(i interface{}) bool
}

type FieldNode struct {
	Field string
	Comp string
	Value string
}


func (f *FieldNode) Matches(i interface{}) bool {
	switch i.(type) {
	case string:
		return i == f.Value
	}
	return false
}

func (f *FieldNode) build(items chan lex.Item) error {
	for i := 0; i < 2; i++ {
		select {
		case item := <-items:
			err := f.set(item)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("item to build FieldNode wasn't available")
		}
	}
	return nil
}

func (f *FieldNode) set(item lex.Item) error {
	switch item.Type {
	case lex.FieldType:
		f.Field = item.Value
	case lex.ComparatorType:
		f.Comp = item.Value
	case lex.LiteralType:
		f.Value = strings.ReplaceAll(item.Value, `"`, "")
	default:
		return fmt.Errorf("invalid type for field node")
	}
	return nil
}

type BooleanNode struct {
	Operation string
	Children []Criteria
}

func (b *BooleanNode) Matches(i interface{}) bool {
	var f func(b1, b2 bool) bool
	state := false
	switch b.Operation{
	case "&&":
		state = true // false dominates 'and' expression
		f = func(b1, b2 bool) bool {return b1 && b2}
	case "||":
		state = false // true dominates 'or' expression
		f = func(b1, b2 bool) bool {return b1 || b2}
	default:
		panic(fmt.Sprintf("%s is an invalid operation", b.Operation))
	}
	for _, c := range b.Children {
		state = f(state, c.Matches(i))
	}
	return state
}

func (b *BooleanNode) build(items chan lex.Item) error {
	for item := range items {
		switch item.Type{
		case lex.BooleanType:
			b.Operation = item.Value
		case lex.OpenPrecedenceType:
			newBool := BooleanNode{}
			err := newBool.build(items)
			if err != nil {
				return err
			}
			b.Children = append(b.Children, &newBool)
		case lex.ClosePrecedenceType:
			return nil
		case lex.LiteralType, lex.FieldType:
			newField := FieldNode{}
			err := newField.set(item)
			if err != nil {
				return err
			}
			err = newField.build(items)
			if err != nil {
				return err
			}
			b.Children = append(b.Children, &newField)
		default:
			return fmt.Errorf("invalid syntax")
		}
	}
	return nil
}

func Parse(items chan lex.Item) (Criteria, error) {
	root := BooleanNode{Operation: "&&"}
	err := root.build(items)
	if err != nil {
		return nil, err
	}
	return &root, nil
}

