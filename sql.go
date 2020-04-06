package funnel

import (
	"fmt"
	"github.com/mchaynes/funnel/internal/squirrel"
	"github.com/mchaynes/funnel/lex"
)

type BoolOp string

const (
	And BoolOp = "AND"
	Or BoolOp = "OR"
	Invalid BoolOp = "INVALID"
)

// ToSql accepts FQL and returns a valid SQL statement
func ToSql(input string, fieldDict []string) (string, []interface{}, error) {
	lexer := lex.NewLexer(input, fieldDict)
	root, err := buildTree(lexer)
	if err != nil {
		return "", nil, err
	}
	sql, args := root.ToSql()
	return sql, args, nil
}

type Format int

const (
	Dollar Format = iota
	Colon
)

func Replace(sql string, format Format) (string, error) {
	switch format{
	case Dollar:
		return squirrel.Dollar.ReplacePlaceholders(sql)
	case Colon:
		return squirrel.Colon.ReplacePlaceholders(sql)
	default:
		panic("illegal format") // caller code is messed up, don't return error, panic
	}
}

type sqler interface {
	ToSql() (string, []interface{})
}


type root struct {
	sqler sqler
}


func (r *root) ToSql() (string, []interface{}) {
	return r.sqler.ToSql()
}

type boolean struct {
	operation BoolOp
	children []sqler
}

func (b *boolean) build(l *lex.Lexer) error {
	for item := l.NextItem(); item != nil; item = l.NextItem(){
		switch item.Type{
		case lex.BooleanType:
			b.operation = toBoolOp(item.Value)
			if b.operation == Invalid {
				return fmt.Errorf("%q is an invalid operation", item.Value)
			}
		case lex.OpenPrecedenceType:
			newBool := boolean{}
			err := newBool.build(l)
			if err != nil {
				return err
			}
			b.children = append(b.children, &newBool)
		case lex.ClosePrecedenceType:
			return nil
		case lex.LiteralType, lex.FieldType:
			newField := field{}
			err := newField.set(*item)
			if err != nil {
				return err
			}
			err = newField.build(l)
			if err != nil {
				return err
			}
			b.children = append(b.children, &newField)
		default:
			return fmt.Errorf("invalid syntax")
		}
	}
	return nil
}

func (f *field) build(l *lex.Lexer) error {
	for i := 0; i < 2; i++ {
		item := l.NextItem()
		if item == nil {
			return fmt.Errorf("unexpected eof")
		}
		if err := f.set(*item); err != nil {
			return err
		}
	}
	return nil
}

func (f *field) set(item lex.Item) error {
	switch item.Type {
	case lex.ComparatorType:
		f.comp = item.Value
	case lex.FieldType:
		f.field = item.Value
	case lex.LiteralType:
		f.val = item.Value
	default:
		return fmt.Errorf("error parsing %v", item.Type)
	}
	return nil
}

func (b *boolean) ToSql() (string, []interface{}) {
	var (
		sql string
		args []interface{}
	)
	sql = "("
	for i, c := range b.children {
		cSql, cArgs := c.ToSql()
		if i > 0 {
			sql = fmt.Sprintf("%s %s %s", sql, b.operation, cSql)
		} else {
			sql = fmt.Sprintf("%s %s", sql, cSql)
		}
		args = append(args, cArgs...)
	}
	sql = fmt.Sprintf("%s )", sql)
	return sql, args
}


type field struct {
	field string
	comp string
	val string
}

func (f *field) ToSql() (string, []interface{}) {
	if len(f.comp) > 2 {
		// ok to panic here, the lexer should have returned an error, so this state is very invalid and indicative of a bug
		panic(fmt.Sprintf("field's comparator was longer than 2 characters: %q", f.comp))
	}
	return fmt.Sprintf("%s %s ?", f.field, f.comp), []interface{}{f.val}
}


func buildTree(l *lex.Lexer) (*root, error) {
	r := &root{}
	b := boolean{}
	err := b.build(l)
	if err != nil {
		return nil, err
	}
	r.sqler = &b
	return r, nil
}


func toBoolOp(b string) BoolOp {
	switch b {
	case string(And):
		return And
	case string(Or):
		return Or
	case lex.Or:
		return Or
	case lex.And:
		return And
	default:
		return Invalid
	}
}