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

func isIdent(t Token) bool {
	return identRegexp.MatchString(t.Text)
}

func scan(r io.Reader) (toks []Token) {
	var s scanner.Scanner
	s.Init(r)
	s.Mode = scanner.ScanIdents | scanner.ScanStrings | scanner.ScanComments
	s.Whitespace = scanner.GoWhitespace
	s.IsIdentRune = func(c rune, i int) bool { return isIdent(Token{Text: string(c)}) }

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

func (p *Parser) eof() bool {
	return p.index >= len(p.toks)
}

func (p *Parser) peek() Token {
	return p.toks[p.index]
}

func (p *Parser) next() {
	p.index++
}

type TokenFnMap map[int]func(t Token) bool

func (p *Parser) parseTokens(when string, m TokenFnMap) (toks []Token, err error) {
	for i := 0; ; i++ {
		f, ok := m[i]
		if !ok {
			return
		}

		if p.eof() {
			err = fmt.Errorf("unexpected EOF when %v", when)
			return
		}

		t := p.peek()

		ok = f(t)
		if !ok {
			err = fmt.Errorf("unexepcted token '%v' when %v [%v:%v]", t.Text, when, t.Line, t.Column)
			return
		}

		toks = append(toks, t)
		p.next()
	}
}

// key=value
type Option struct {
	Left, Right Token
}

func (p *Parser) parseOption() (opt *Option, err error) {
	toks, err := p.parseTokens("parseOption", TokenFnMap{
		0: isIdent,
		1: func(t Token) bool { return t.Text == "=" },
		2: isIdent,
	})

	if err != nil {
		return nil, err
	} else {
		return &Option{
			Left:  toks[0],
			Right: toks[2],
		}, nil
	}
}

// tok tok tok ... max to 5
type Idents struct {
	Token []Token
}

func (p *Parser) parseIdents() (idents *Idents, err error) {
	toks, _ := p.parseTokens("parseIdents", TokenFnMap{
		0: isIdent,
		1: isIdent,
		2: isIdent,
		3: isIdent,
		4: isIdent,
		5: isIdent,
	})

	return &Idents{
		Token: toks,
	}, nil
}

func TestA(t *testing.T) {
	toks := scan(bytes.NewReader([]byte("abc def ")))
	fmt.Println(toks)
	p := NewParser(toks)
	fmt.Println(p.parseIdents())
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
