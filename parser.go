package main

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"text/scanner"
)

type Token struct {
	Text   string
	Line   int
	Column int
}

var identRegex = regexp.MustCompile(`^[a-zA-Z0-9\.\-_\(\)]+$`)

func isIdent(s string) bool {
	return identRegex.MatchString(s)
}

type Parser struct {
	toks []*Token
	pos  int

	context []string
}

func (p *Parser) Parse(r io.Reader) (b *Block, err error) {
	var s scanner.Scanner

	s.Init(r)
	s.Mode = scanner.ScanIdents | scanner.ScanStrings | scanner.ScanComments
	s.Whitespace = scanner.GoWhitespace
	s.IsIdentRune = func(c rune, i int) bool { return isIdent(string(c)) }

	for {
		if t := s.Scan(); t != scanner.EOF {
			p.toks = append(p.toks, &Token{
				Text:   s.TokenText(),
				Line:   s.Position.Line,
				Column: s.Position.Column,
			})
		} else {
			break
		}
	}

	defer func() {
		if msg := recover(); msg != nil {
			err = fmt.Errorf("parse failed: %v", msg)
		}
	}()

	return p.mustParseBlock(), nil
}

func (p *Parser) eof() bool {
	return p.pos >= len(p.toks)
}

func (p *Parser) peek() *Token {
	if p.eof() {
		return nil
	}
	return p.toks[p.pos]
}

func (p *Parser) next() *Token {
	t := p.peek()
	p.pos++
	return t
}

func (p *Parser) pushContext(s string) {
	p.context = append(p.context, s)
}

func (p *Parser) popContext() {
	p.context = p.context[:len(p.context)-1]
}

func (p *Parser) contextPath() string {
	var s string
	for _, c := range p.context {
		s = s + "." + c
	}
	return s
}

func (p *Parser) COMPLAIN() {
	if p.eof() {
		panic(fmt.Sprintf("unexpected EOF when parse:[%v]", p.contextPath()))
	} else {
		t := p.peek()
		panic(fmt.Sprintf("unexpected token '%v' when parse:[%v], position:[%v:%v]",
			t.Text, p.contextPath(), t.Line, t.Column))
	}
}

func (p *Parser) checkTokenIf(f TokenCheckFn) bool {
	return !p.eof() && f(p.peek())
}

func (p *Parser) mustParseToken(f TokenCheckFn) *Token {
	if !p.checkTokenIf(f) {
		p.COMPLAIN()
	}

	return p.next()
}

type TokenCheckFn func(t *Token) bool

func checkIfIdentifer(t *Token) bool      { return isIdent(t.Text) }
func checkIfEquals(s string) TokenCheckFn { return func(t *Token) bool { return t.Text == s } }
func checkIfComment(t *Token) bool        { return strings.HasPrefix(t.Text, "//") }

// Identifer can be either a keyword | string | number | variable
type Identifer struct {
	T *Token
}

func (p *Parser) mustParseIdentifer() *Identifer {
	p.pushContext("identifer")
	defer p.popContext()

	return &Identifer{p.mustParseToken(checkIfIdentifer)}
}

// Instruction is a list of Identifers (number > 0)
type Instruction struct {
	I []*Identifer
}

func (p *Parser) mustParseInstruction() *Instruction {
	p.pushContext("instruction")
	defer p.popContext()

	ins := &Instruction{}

	for {
		if p.checkTokenIf(checkIfIdentifer) {
			ins.I = append(ins.I, p.mustParseIdentifer())
			continue
		}

		if len(ins.I) > 0 {
			return ins
		}

		p.COMPLAIN()
	}
}

// Value has 2 forms:
//   1. Identifer
//   2. Identifer [Option = OptionValue]
type Value struct {
	I    *Identifer
	K, V *Identifer
}

func (p *Parser) mustParseValue() *Value {
	p.pushContext("value")
	defer p.popContext()

	v := &Value{}

	v.I = p.mustParseIdentifer()

	if !p.checkTokenIf(checkIfEquals("[")) { // form1
		return v
	}

	// form2
	p.pushContext("option")
	defer p.popContext()

	p.mustParseToken(checkIfEquals("["))
	defer p.mustParseToken(checkIfEquals("]"))

	v.K = p.mustParseIdentifer()
	p.mustParseToken(checkIfEquals("="))
	v.V = p.mustParseIdentifer()

	return v
}

// Statement has 3 forms:
//   1. Instruction;
//   2. Instruction = Value;
//   3. Instruction { Block }
//   4. Comment
type Statement struct {
	I *Instruction
	V *Value
	B *Block
	C *Token
}

func (p *Parser) mustParseStatement() *Statement {
	p.pushContext("statement")
	defer p.popContext()

	s := &Statement{}

	if p.checkTokenIf(checkIfComment) { // form4
		s.C = p.next()
		return s
	}

	s.I = p.mustParseInstruction()

	if p.checkTokenIf(checkIfEquals(";")) { // form1
		p.next()

		return s
	}

	if p.checkTokenIf(checkIfEquals("=")) { // form2
		p.next()

		s.V = p.mustParseValue()
		p.mustParseToken(checkIfEquals(";"))

		return s
	}

	if p.checkTokenIf(checkIfEquals("{")) { // form3
		p.next()

		s.B = p.mustParseBlock()
		p.mustParseToken(checkIfEquals("}"))

		return s
	}

	p.COMPLAIN()
	return nil
}

// Block is a list of Statements (number >= 0)
type Block struct {
	S []*Statement
}

func (p *Parser) mustParseBlock() *Block {
	p.pushContext("block")
	defer p.popContext()

	b := &Block{}
	for {
		if p.eof() || p.checkTokenIf(checkIfEquals("}")) {
			break
		}

		b.S = append(b.S, p.mustParseStatement())
	}

	return b
}
