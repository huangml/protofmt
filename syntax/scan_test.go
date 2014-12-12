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

var text = `
hello321.456 world

message haha 123 {
	optional int32 x = 100[abc=def]; // fdas fdsa fdsa
}

// abc
// def

set x = "string";
`

func TestA(t *testing.T) {
	for _, v := range scan(bytes.NewReader([]byte(text))) {
		fmt.Println(v)
	}
}

func wordRune(ch rune, i int) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' ||
		ch == '_' || ch == '-' || ch == '.'
}

var regex = regexp.MustCompile(`[a-zA-Z0-9.-_]`)

func scan(r io.Reader) (toks []Token) {
	var s scanner.Scanner
	s.Init(r)
	s.Mode = scanner.ScanIdents | scanner.ScanStrings | scanner.ScanComments
	s.Whitespace = scanner.GoWhitespace
	s.IsIdentRune = wordRune

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

// key=value
type Option struct {
	Left, Right Token
}

// tok1 tok2;
type Declare struct {
	Token []Token
}

// left1 left2 = right1 right2 [key = value];
type Equation struct {
	Left, Right []Token
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
