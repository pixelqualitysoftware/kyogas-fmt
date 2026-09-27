package formatter

import (
	"fmt"
	"os"
)

func Format(file string) {
	l, err := NewLexer(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Scan the file
	toks := l.Scan()
	for _, tok := range toks {
		fmt.Println(tok)
	}

	fmt.Println()

	p := NewParser(file, toks)
	ast := p.Parse()

	ast.Print(0)
}
