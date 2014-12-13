package syntax

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"testing"
	"text/scanner"
)

type Token struct {
	Text   string
	Line   int
	Column int
}

var identRegexp = regexp.MustCompile(`[a-zA-Z0-9\.\-_\(\)]`)

func isIdent(s string) bool {
	return identRegexp.MatchString(s)
}

var text = `
default=}100
`

func TestA(t *testing.T) {
	toks := scan(bytes.NewReader([]byte(text)))
	p := NewParser(toks)
	fmt.Println(p.parseOption())
}

func scan(r io.Reader) (toks []Token) {
	var s scanner.Scanner
	s.Init(r)
	s.Mode = scanner.ScanIdents | scanner.ScanStrings | scanner.ScanComments
	s.Whitespace = scanner.GoWhitespace
	s.IsIdentRune = func(c rune, i int) bool { return isIdent(string(c)) }

	for {
		if t := s.Scan(); t != scanner.EOF {
			toks = append(toks, Token{
				Text:   s.TokenText(),
				Line:   s.Position.Line,
				Column: s.Position.Column,
			})
		} else {
			return
		}
	}
}

type Parser struct {
	toks  []Token
	index int
}

func NewParser(toks []Token) *Parser {
	return &Parser{
		toks: toks,
	}
}

func (p *Parser) EOF() bool {
	return p.index >= len(p.toks)
}

func (p *Parser) peek() Token {
	return p.toks[p.index]
}

func (p *Parser) next() Token {
	defer func() { p.index++ }()
	return p.toks[p.index]
}

// key=value
type Option struct {
	Left, Right Token
}

func (p *Parser) parseOption() (opt *Option, err error) {
	when := "parseOption"
	checkEOF := func() bool {
		if p.EOF() {
			opt, err = nil, EOFError(when)
			return false
		}
		return true
	}

	var t [3]Token
	for i := range t {
		if !checkEOF() {
			return
		}
		t[i] = p.next()
	}

	if !isIdent(t[0].Text) {
		return nil, TokenError(when, t[0])
	}

	if t[1].Text != "=" {
		return nil, TokenError(when, t[1])
	}

	if !isIdent(t[2].Text) {
		return nil, TokenError(when, t[2])
	}

	return &Option{
		Left:  t[0],
		Right: t[2],
	}, nil
}

func EOFError(when string) error {
	return fmt.Errorf("unexpected EOF when %v", when)
}

func TokenError(when string, tok Token) error {
	return fmt.Errorf("unexepcted token '%v' when %v [%v:%v]", tok.Text, when, tok.Line, tok.Column)
}

// tok1 tok2
type Declare struct {
	Token []Token
}

// left1 left2 = right1[key=value]
type Equation struct {
	Left  []Token
	Right Token
	*Option
}

// tok1 tok2 { Block }
type Structure struct {
	Token []Token
	Group []Group
}

type Group struct {
	Statements []interface{}
}
