package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestA(t *testing.T) {
	text := `
	message hello {
	     message world {
	             optional int32 i = 1[default=100];
	     }

	     repeated world w = 1;
	     optional int32 i = 2;
	}
	`

	var p Parser
	b, err := p.Parse(bytes.NewReader([]byte(text)))
	if err != nil {
		fmt.Println(err)
		return
	}

	var f Formatter
	f.writeBlock(b)
	f.print()
}
