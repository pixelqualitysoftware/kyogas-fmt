package formatter

import (
	"fmt"
	"os"
	"strings"
)

type NodeType int

const (
	NodeTypeVariableDef NodeType = iota
	NodeTypeArrayDef
	NodeTypeDictDef

	// Types
	NodeTypeStr
	NodeTypeFloat
	NodeTypeNum // We don't really need uint, int, etc since our job is just to format the file

	// Compiler stuff
	NodeTypeRoot
)

type Node struct {
	Type   NodeType
	Left   *Node
	Right  *Node
	Lexeme string

	Value any
}

type NodeVariableDef struct {
	Type  NodeType
	Value Node
}

type NodeArrayDef struct {
	Elements []Node
}

type NodeDictDef struct {
	Dict map[string]Node
}

type NodeBlock NodeArrayDef

func (nt NodeType) String() string {
	switch nt {
	case NodeTypeVariableDef:
		return "variable def"
	case NodeTypeArrayDef:
		return "array def"
	case NodeTypeDictDef:
		return "dict def"
	case NodeTypeStr:
		return "str"
	case NodeTypeFloat:
		return "float"
	case NodeTypeNum:
		return "num"
	case NodeTypeRoot:
		return "root"
	default:
		return "unknown"
	}
}

func (n *Node) Print(depth int) {
	if n == nil {
		return
	}

	pad := strings.Repeat("  ", depth)

	// Print node's type and lexeme
	if n.Lexeme != "" {
		fmt.Printf("%s%s (%q)", pad, n.Type, n.Lexeme)
	} else {
		fmt.Printf("%s%s", pad, n.Type)
	}

	// Print value
	if n.Value != nil {
		switch v := n.Value.(type) {
		case NodeVariableDef:
			fmt.Printf(" -> (type: %s) value:\n", v.Type)
			v.Value.Print(depth + 1)

		case NodeArrayDef:
			fmt.Println(" -> array:")
			for i, elem := range v.Elements {
				fmt.Printf("%s  [%d]:\n", pad, i)
				elem.Print(depth + 2)
			}

		case NodeBlock:
			fmt.Println(" -> block:")
			for i, elem := range v.Elements {
				fmt.Printf("%s  [%d]:\n", pad, i)
				elem.Print(depth + 2)
			}

		case *NodeBlock:
			fmt.Println(" -> block:")
			if v != nil {
				for i, elem := range v.Elements {
					fmt.Printf("%s  [%d]:\n", pad, i)
					elem.Print(depth + 2)
				}
			}

		case NodeDictDef:
			fmt.Println(" -> dict:")
			for key, val := range v.Dict {
				fmt.Printf("%s  key %v:\n", pad, key)
				val.Print(depth + 2)
			}

		default:
			fmt.Printf(" -> value: %v\n", v)
		}
	} else {
		fmt.Println()
	}

	// Print left
	if n.Left != nil {
		fmt.Printf("%s  left:\n", pad)
		n.Left.Print(depth + 2)
	}

	// Print right
	if n.Right != nil {
		fmt.Printf("%s  right:\n", pad)
		n.Right.Print(depth + 2)
	}
}

func (n *Node) AppendElement(element Node) {
	switch v := n.Value.(type) {
	case NodeArrayDef:
		v.Elements = append(v.Elements, element)
		n.Value = v

	case *NodeArrayDef:
		v.Elements = append(v.Elements, element)

	case NodeBlock:
		v.Elements = append(v.Elements, element)
		n.Value = v

	case *NodeBlock:
		v.Elements = append(v.Elements, element)
	}
}

type Parser struct {
	pos  int
	file string
	toks []Token
}

func NewParser(file string, toks []Token) *Parser {
	return &Parser{
		pos:  0,
		file: file,
		toks: toks,
	}
}

func (p *Parser) peek() Token {
	if p.pos >= len(p.toks) {
		return p.toks[len(p.toks)-1]
	}

	return p.toks[p.pos]
}

func (p *Parser) consume() Token {
	t := p.toks[p.pos]
	p.pos++

	return t
}

func (p *Parser) isType(t Token) bool {
	switch t.Type {
	case TokenBool, TokenInt, TokenFlt, TokenStr, TokenUint,
		TokenByte, TokenArr, TokenObj, TokenShort, TokenUShort,
		TokenSByte, TokenLong, TokenULong:
		return true
	default:
		return false
	}
}

func (p *Parser) expect(t TokenType) error {
	tok := p.peek()
	if tok.Type != t {
		return fmt.Errorf("%s:%d:%d: expected %s, got %s", p.file, tok.Line, tok.Column, t.String(), tok.Type.String())
	}

	return nil
}

func typeToNode(tokType TokenType) NodeType {
	switch tokType {
	case TokenStr, TokenStrLit:
		return NodeTypeStr

	case TokenFlt, TokenFloatLit:
		return NodeTypeFloat

	case TokenInt, TokenUint, TokenByte, TokenShort, TokenUShort, TokenSByte, TokenLong, TokenULong, TokenNumLit:
		return NodeTypeNum

	default:
		return NodeTypeNum
	}
}

func (p *Parser) parseVar() (*Node, error) {
	varType := p.consume() // Save the type for later

	err := p.expect(TokenLit)
	if err != nil {
		return nil, err
	}

	name := p.consume() // Save the name too

	err = p.expect(TokenColon)
	if err != nil {
		return nil, err
	}

	p.consume()

	t := p.peek()
	if t.Type == TokenEof {
		return nil, fmt.Errorf("%s:%d:%d: expected a value, got %s", p.file, t.Line, t.Column, t.Type.String())
	}

	p.consume()

	temp := Node{
		Lexeme: t.Value,
		Type:   typeToNode(t.Type),
	}

	val := NodeVariableDef{
		Type:  typeToNode(varType.Type),
		Value: temp,
	}

	return &Node{
		Lexeme: name.Value,
		Type:   NodeTypeVariableDef,
		Value:  val,
	}, nil
}

func (p *Parser) Parse() Node {
	var ast = Node{
		Type:  NodeTypeRoot,
		Value: NodeBlock{},
	}
	for {
		t := p.peek()
		if t.Type == TokenEof {
			break // End of tokens
		}

		if p.isType(t) {
			// Variable def
			n, err := p.parseVar()
			if err != nil {
				fmt.Printf("kyogas-fmt: error: %v\n", err)
				os.Exit(1)
			}

			ast.AppendElement(*n)
		}
	}

	return ast
}
