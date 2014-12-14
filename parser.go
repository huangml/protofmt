package main

import (
	"fmt"
	"io"
	"regexp"
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

func (p *Parser) scan(r io.Reader) {

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

func (p *Parser) checkTokenIf(f TokenCheckFn) bool {
	return !p.eof() && f(p.peek())
}

func (p *Parser) parseTokenIf(f TokenCheckFn) (*Token, error) {
	if p.eof() {
		return nil, fmt.Errorf("unexpected EOF when parse:[%v]", p.contextPath())
	}

	t := p.peek()
	if !f(t) {
		return nil, fmt.Errorf("unexpected token '%v' when parse:[%v], position:[%v:%v]", t.Text, p.contextPath(), t.Line, t.Column)
	}

	p.next()
	return t, nil
}

type TokenCheckFn func(t *Token) bool

func checkIfIdentifer(t *Token) bool      { return isIdent(t.Text) }
func checkIfEquals(s string) TokenCheckFn { return func(t *Token) bool { return t.Text == s } }

// Identifer can be either a keyword | string | number | variable
type Identifer struct {
	T *Token
}

func (p *Parser) parseIdentifer() (*Identifer, error) {
	p.pushContext("identifer")
	defer p.popContext()

	if tok, err := p.parseTokenIf(checkIfIdentifer); err != nil {
		return nil, err
	} else {
		return &Identifer{tok}, nil
	}
}

// Instruction is a list of Identifers
type Instruction struct {
	I []*Identifer
}

func (p *Parser) parseInstruction() (*Instruction, error) {
	p.pushContext("instruction")
	defer p.popContext()

	var ins Instruction

	for {
		if I, err := p.parseIdentifer(); err == nil {
			ins.I = append(ins.I, I)
		} else if len(ins.I) > 0 {
			return &ins, nil
		} else {
			return nil, err
		}
	}
}

// Value has 2 forms:
//   1. Identifer
//   2. Identifer [Option = OptionValue]
type Value struct {
	I, K, V *Identifer
}

func (p *Parser) parseValue() (*Value, error) {
	p.pushContext("value")
	defer p.popContext()

	I, err := p.parseIdentifer()
	if err != nil {
		return nil, err
	}

	if !p.checkTokenIf(checkIfEquals("[")) { // form1
		return &Value{I: I}, nil
	}
	p.next()

	// form2
	p.pushContext("option")
	defer p.popContext()

	K, err := p.parseIdentifer()
	if err != nil {
		return nil, err
	}

	if _, err := p.parseTokenIf(checkIfEquals("=")); err != nil {
		return nil, err
	}

	V, err := p.parseIdentifer()
	if err != nil {
		return nil, err
	}

	if _, err := p.parseTokenIf(checkIfEquals("]")); err != nil {
		return nil, err
	}

	return &Value{I: I, K: K, V: V}, nil
}

// Statement has 3 forms:
//   1. Instruction;
//   2. Instruction = Value;
//   3. Instruction { Block }
type Statement struct {
	I *Instruction
	V *Value
	B *Block
}

func (p *Parser) parseStatement() (*Statement, error) {
	p.pushContext("statement")
	defer p.popContext()

	I, err := p.parseInstruction()
	if err != nil {
		return nil, err
	}

	if p.checkTokenIf(checkIfEquals(";")) { // form1
		p.next()

		return &Statement{I: I}, nil
	}

	if p.checkTokenIf(checkIfEquals("=")) { // form2
		p.next()

		V, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		if _, err = p.parseTokenIf(checkIfEquals(";")); err != nil {
			return nil, err
		}

		return &Statement{I: I, V: V}, nil
	}

	if p.checkTokenIf(checkIfEquals("{")) { // form3
		p.next()

		B, err := p.parseBlock()
		if err != nil {
			return nil, err
		}

		if _, err = p.parseTokenIf(checkIfEquals("}")); err != nil {
			return nil, err
		}

		return &Statement{I: I, B: B}, nil
	}

	// always get an error, since we already checked form1
	_, err = p.parseTokenIf(checkIfEquals(";"))
	return nil, err
}

// Block is a list of Statements
type Block struct {
	S []*Statement
}

func (p *Parser) parseBlock() (*Block, error) {
	p.pushContext("block")
	defer p.popContext()

	var b Block
	for {
		if p.eof() || p.checkTokenIf(checkIfEquals("}")) {
			break
		}

		if S, err := p.parseStatement(); err != nil {
			return nil, err
		} else {
			b.S = append(b.S, S)
		}
	}

	return &b, nil
}
