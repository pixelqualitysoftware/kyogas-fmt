package formatter

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type Lexer struct {
	line, column, errors, pos int

	file    string
	fileCtx []byte

	buffer  strings.Builder
	types   map[string]TokenType
	symbols map[string]TokenType
}

type TokenType int

const (
	// Compiler stuff
	TokenStrLit TokenType = iota
	TokenLit
	TokenNumLit
	TokenFloatLit

	TokenEof

	// Symbols
	TokenLessThan
	TokenGreaterThan
	TokenMinus
	TokenEquals

	TokenArrowLeft
	TokenArrowRight

	TokenDArrowLeft
	TokenDArrowRight
	TokenColon

	// Types
	TokenBool
	TokenInt
	TokenFlt
	TokenStr
	TokenUint
	TokenByte
	TokenArr
	TokenObj

	TokenShort
	TokenUShort
	TokenSByte
	TokenLong
	TokenULong
)

type Token struct {
	Type         TokenType
	Line, Column int
	Value        string
}

func (t TokenType) String() string {
	switch t {
	case TokenStrLit:
		return "str lit"
	case TokenLit:
		return "lit"
	case TokenNumLit:
		return "num lit"
	case TokenFloatLit:
		return "float lit"
	case TokenLessThan:
		return "<"
	case TokenGreaterThan:
		return ">"
	case TokenMinus:
		return "-"
	case TokenEquals:
		return "="
	case TokenArrowLeft:
		return "<-"
	case TokenArrowRight:
		return "->"
	case TokenDArrowLeft:
		return "<=="
	case TokenDArrowRight:
		return "==>"
	case TokenColon:
		return ":"
	case TokenBool:
		return "bool"
	case TokenInt:
		return "int"
	case TokenFlt:
		return "flt"
	case TokenStr:
		return "str"
	case TokenUint:
		return "uint"
	case TokenByte:
		return "byte"
	case TokenArr:
		return "arr"
	case TokenObj:
		return "obj"
	case TokenShort:
		return "short"
	case TokenUShort:
		return "ushort"
	case TokenSByte:
		return "sbyte"
	case TokenLong:
		return "long"
	case TokenULong:
		return "ulong"
	case TokenEof:
		return "<eof>"
	default:
		return "unknown"
	}
}

func (t Token) String() string {
	if t.Value == "" {
		return fmt.Sprintf("%s at %d:%d", t.Type, t.Line, t.Column)
	}
	return fmt.Sprintf("%s containing %q at %d:%d", t.Type, t.Value, t.Line, t.Column)
}

func NewLexer(file string) (*Lexer, error) {
	dat, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return &Lexer{
		line:   1,
		column: 1,
		pos:    0,

		file:    file,
		fileCtx: dat,
		buffer:  strings.Builder{},

		types: map[string]TokenType{
			"bool":   TokenBool,
			"int":    TokenInt,
			"flt":    TokenFlt,
			"str":    TokenStr,
			"uint":   TokenUint,
			"byte":   TokenByte,
			"arr":    TokenArr,
			"obj":    TokenObj,
			"short":  TokenShort,
			"ushort": TokenUShort,
			"sbyte":  TokenSByte,
			"long":   TokenLong,
			"ulong":  TokenULong,
		},

		symbols: map[string]TokenType{
			"->":  TokenArrowRight,
			"<-":  TokenArrowLeft,
			"==>": TokenDArrowRight,
			"<==": TokenDArrowLeft,
			"<":   TokenLessThan,
			">":   TokenGreaterThan,
			"-":   TokenMinus,
			"=":   TokenEquals,
			":":   TokenColon,
		},
	}, nil
}

func (l *Lexer) consume() byte {
	c := l.fileCtx[l.pos]
	l.pos++

	if c == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}

	return c
}

func (l *Lexer) peek() (byte, bool) {
	if l.pos >= len(l.fileCtx) {
		return 0, false
	}

	return l.fileCtx[l.pos], true
}

func (l *Lexer) findType(strType string) (TokenType, bool) {
	t, ok := l.types[strType]
	return t, ok
}

func (l *Lexer) findSymbol(sym string) (TokenType, bool) {
	t, ok := l.symbols[sym]
	return t, ok
}

func (l *Lexer) isSymbol(c byte) bool {
	if c == '-' || c == '=' || c == '<' || c == '>' || c == ':' {
		return true
	} else {
		return false
	}
}

func (l *Lexer) Scan() []Token {
	var toks []Token
	for {
		// Get current character
		c, ok := l.peek()
		if !ok {
			break // We hit EOF
		}

		// Is a letter. Maybe it's a type?
		if unicode.IsLetter(rune(c)) {
			startL, startC := l.line, l.column

			for {
				c, ok := l.peek()

				if !ok || !unicode.IsLetter(rune(c)) && c != '-' && !unicode.IsDigit(rune(c)) {
					break
				}

				l.buffer.WriteByte(c)
				l.consume()
			}

			t, ok := l.findType(l.buffer.String())
			if !ok {
				// It's a literal
				tok := Token{
					Line:   startL,
					Column: startC,
					Type:   TokenLit,
					Value:  l.buffer.String(),
				}

				toks = append(toks, tok)
				l.buffer.Reset()
			} else {
				// It's a type
				tok := Token{
					Line:   startL,
					Column: startC,
					Type:   t,
				}

				toks = append(toks, tok)
				l.buffer.Reset()
			}
		} else if unicode.IsDigit(rune(c)) {
			startL, startC := l.line, l.column
			for {
				c, ok := l.peek()

				// Check also if c != . because it may be a float
				if !ok || !unicode.IsDigit(rune(c)) && c != '.' {
					break
				}

				l.buffer.WriteByte(c)
				l.consume()
			}

			// It's a float
			if strings.Contains(l.buffer.String(), ".") {
				_, err := strconv.ParseFloat(l.buffer.String(), 64) // We don't really need the real float, we just check if it's formatted correctly
				if err != nil {
					fmt.Printf("kyogas-fmt: %s:%d:%d: %v\n", l.file, startL, startC, err)
					os.Exit(1)
				}

				tok := Token{
					Line:   startL,
					Column: startC,
					Type:   TokenFloatLit,
					Value:  l.buffer.String(),
				}

				toks = append(toks, tok)
				l.buffer.Reset()
			} else {
				// It's an int
				_, err := strconv.ParseInt(l.buffer.String(), 10, 64)
				if err != nil {
					fmt.Printf("kyogas-fmt: %s:%d:%d: %v\n", l.file, startL, startC, err)
					os.Exit(1)
				}

				tok := Token{
					Line:   startL,
					Column: startC,
					Type:   TokenNumLit,
					Value:  l.buffer.String(),
				}

				toks = append(toks, tok)
				l.buffer.Reset()
			}
		} else if unicode.IsSpace(rune(c)) {
			// Handles spaces, tabs and newlines
			l.consume()
		} else if l.isSymbol(c) {
			// Handle symbols like <-
			startL, startC := l.line, l.column
			for {
				c, ok := l.peek()

				if !ok || !l.isSymbol(c) {
					break
				}

				l.buffer.WriteByte(c)
				l.consume()
			}

			s, ok := l.findSymbol(l.buffer.String())
			if !ok {
				fmt.Printf("kyogas-fmt: %s:%d:%d: unknown symbol %q\n", l.file, startL, startC, l.buffer.String())
				os.Exit(1)
			}

			tok := Token{
				Line:   startL,
				Column: startC,
				Type:   s,
			}

			toks = append(toks, tok)
			l.buffer.Reset()
		} else if c == '|' {
			// Ignore everything until a newline
			for {
				c, ok := l.peek()

				if !ok || c == '\n' {
					break
				}

				l.consume()
			}
		} else if c == '"' {
			// Handle string
			startL, startC := l.line, l.column
			l.consume()

			for {
				c, ok := l.peek()

				if !ok || c == '"' {
					break
				}

				l.buffer.WriteByte(c)
				l.consume()
			}

			l.consume()
			tok := Token{
				Line:   startL,
				Column: startC,
				Type:   TokenStrLit,
				Value:  l.buffer.String(),
			}

			toks = append(toks, tok)
			l.buffer.Reset()
		} else {
			fmt.Printf("kyogas-fmt: %s:%d:%d: unknown character %q\n", l.file, l.line, l.column, string(c))
			os.Exit(1)
		}
	}

	toks = append(toks, Token{
		Line:   l.line,
		Column: l.column,
		Type:   TokenEof,
	})

	return toks
}
