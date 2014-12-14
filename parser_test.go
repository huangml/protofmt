package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func TestA(t *testing.T) {

	text := ""
	//      text := `
	// message hello {
	//      message world {
	//              optional int32 i = 1[default=100];
	//      }

	//      repeated world w = 1;
	//      optional int32 i = 2;
	// }
	//      `

	text = "message world {a{}}"

	var p Parser
	p.scan(bytes.NewReader([]byte(text)))

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("ERROR: ", err)
		}
	}()

	s := p.mustParseStatement()

	b, _ := json.MarshalIndent(s, "", "  ")
	fmt.Println(string(b))
}
