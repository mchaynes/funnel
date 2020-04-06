package lex

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	Or  = "||"
	And = "&&"
	Lt  = "<"
	Lte = "<="
	Eq  = "="
	Gt  = ">"
	Gte = ">="
)

const (
	eof          rune = utf8.RuneError
	upperalpha        = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	loweralpha        = "abcdefghijklmopqrstuvwxyz"
	digits            = "0123456789"
	alphanumeric      = upperalpha + loweralpha + digits
)

type Item struct {
	Type  ItemType
	Value string
}

type ItemType int

const (
	ErrorType ItemType = iota
	OpenPrecedenceType
	ClosePrecedenceType
	BooleanType
	LiteralType
	FieldType
	ComparatorType
)

const (
	openPrecedence  = '('
	closePrecedence = ')'
	quoteLiteral    = '"'
	and             = '&'
	or              = '|'
)

type stateFunc func(lexer *Lexer) stateFunc

type Lexer struct {
	input    string
	position int
	width    int
	start    int
	items    chan Item
	state    stateFunc
	dict     map[string]struct{}
}


func NewLexer(input string, allowedFields []string) *Lexer {

	return &Lexer{
		input: input,
		items: make(chan Item, 2),
		state: lexText,
		dict:  buildDict(allowedFields),
	}
}

func (l *Lexer) NextItem() *Item {
	for {
		select {
		case item := <-l.items:
			return &item
		default:
			if l.state == nil {
				return nil
			}
			l.state = l.state(l)
		}
	}
}

func (l *Lexer) next() (r rune) {
	if l.position > len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.position:])
	l.position += l.width
	return r
}

func (l *Lexer) peek() (r rune) {
	r = l.next()
	l.backup()
	return r
}

func (l *Lexer) ignore() {
	l.start = l.position
}

func (l *Lexer) backup() {
	l.position -= l.width
}

func (l *Lexer) emit(itemType ItemType) {
	buf := l.buf()
	if itemType == LiteralType {
		buf = strings.ReplaceAll(buf, `"`, "")
	}
	l.items <- Item{
		Type:  itemType,
		Value: buf,
	}
	l.start = l.position
}

func (l *Lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *Lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {}
	l.backup()
}

func (l *Lexer) acceptUntil(r ...rune) {
	for n := l.next(); !inRunes(n, append(r, eof)...); n = l.next() {}
}

func (l *Lexer) errorf(stmt string, args ...interface{}) stateFunc {
	l.items <- Item{
		Type:  ErrorType,
		Value: fmt.Sprintf(stmt, args...),
	}
	return nil
}

func (l *Lexer) buf() string {
	return l.input[l.start:l.position]
}

func lexText(l *Lexer) stateFunc {
	for {
		switch l.next() {
		case eof:
			return nil
		case openPrecedence:
			return inOpenPrecedence(l)
		case closePrecedence:
			return inClosePrecedence(l)
		case quoteLiteral:
			return inLiteral(l)
		case and, or:
			return inBoolean(l)
		case '<', '>', '=', '!':
			return inComp(l)
		case ' ':
			l.ignore()
		default:
			return inKeyDef(l, l.dict)
		}
	}
}

func inOpenPrecedence(l *Lexer) stateFunc {
	l.emit(OpenPrecedenceType)
	return lexText
}

func inClosePrecedence(l *Lexer) stateFunc {
	l.emit(ClosePrecedenceType)
	return lexText
}

func inLiteral(l *Lexer) stateFunc {
	l.acceptUntil('"')
	l.emit(LiteralType)
	return lexText
}

func inKeyDef(l *Lexer, dict map[string]struct{}) stateFunc {
	l.acceptRun(alphanumeric + ".")
	buf := l.buf()
	// Ensure that we properly detect illegal fields during lexing, protect against illegal filtering/sql injection
	if _, ok := dict[buf]; !ok {
		return l.errorf("%q isn't a valid field", buf)
	}
	l.emit(FieldType)
	return lexText
}

func inBoolean(l *Lexer) stateFunc {
	l.backup()
	f := l.next()
	s := l.next()
	if f != s {
		return l.errorf("invalid boolean expression, %c%c", f, s)
	}
	l.emit(BooleanType)
	return lexText
}

func inComp(l *Lexer) stateFunc {
	l.backup()
	n := fmt.Sprintf("%c%c", l.next(), l.peek())
	p := l.peek()
	// every comparator that has more than one rune ends with '='. ex: '!=', '<=', ">=
	if p == '=' {
		l.next()
	} else {
		n = n[:1]
	}
	l.emit(ComparatorType)
	return lexText
}

func inRunes(r rune, rs ...rune) bool {
	for _, c := range rs {
		if c == r {
			return true
		}
	}
	return false
}

func buildDict(strs []string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, s := range strs {
		m[s] = struct{}{}
	}
	return m
}
