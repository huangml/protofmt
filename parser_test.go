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

	text = `
        message world {
        }
        `

	var p Parser
	p.scan(bytes.NewReader([]byte(text)))

	i, err := p.parseStatement()
	if err == nil {
		b, _ := json.MarshalIndent(i, "", "  ")
		fmt.Println(string(b))
	} else {
		fmt.Println("error: ", err)
	}
}
