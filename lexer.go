package main

import "unicode"

const EOF = '\x04'

type TokenKind int

const (
	TokenKindIdent TokenKind = iota
	TokenKindLabel
	TokenKindInt
	TokenKindInput
	TokenKindOutput
	TokenKindGoto
	TokenKindIf
	TokenKindThen
	TokenKindEqual
	TokenKindPlus
	TokenKindLessThan
	TokenKindInvalid
	TokenKindEnd
)

var tokenKindName = map[TokenKind]string{
	TokenKindIdent:    "ident",
	TokenKindLabel:    "label",
	TokenKindInt:      "int",
	TokenKindInput:    "input",
	TokenKindOutput:   "output",
	TokenKindGoto:     "goto",
	TokenKindIf:       "if",
	TokenKindThen:     "then",
	TokenKindEqual:    "equal",
	TokenKindPlus:     "plus",
	TokenKindLessThan: "lessthan",
	TokenKindInvalid:  "invalid",
	TokenKindEnd:      "end",
}

func (kind TokenKind) String() string {
	return tokenKindName[kind]
}

type Token struct {
	kind TokenKind
	val  string
}

func (t Token) String() string {
	kindStr := t.kind.String()
	var valStr string
	if t.val != "" {
		valStr = "(" + t.val + ")"
	}
	return kindStr + valStr
}

type Lexer struct {
	buf      []byte
	pos      int
	read_pos int
	ch       byte
}

func (l *Lexer) peek() byte {
	if l.read_pos >= len(l.buf) {
		return EOF
	}

	return l.buf[l.read_pos]
}

func (l *Lexer) read() byte {
	l.ch = l.peek()
	l.pos = l.read_pos
	l.read_pos++

	return l.ch
}

func (l *Lexer) nextToken() Token {
	l.skipWhitespace()

	switch {
	case l.ch == EOF:
		l.read()
		return Token{kind: TokenKindEnd}
	case l.ch == '=':
		l.read()
		return Token{kind: TokenKindEqual}
	case l.ch == '+':
		l.read()
		return Token{kind: TokenKindPlus}
	case l.ch == '<':
		l.read()
		return Token{kind: TokenKindLessThan}
	case l.ch == ':':
		l.read()
		p := l.pos
		sl := 0
		for unicode.IsLetter(rune(l.ch)) || l.ch == '_' {
			sl++
			l.read()
		}
		s := l.buf[p : p+sl]
		return Token{kind: TokenKindLabel, val: string(s)}
	case unicode.IsDigit(rune(l.ch)):
		p := l.pos
		sl := 0
		for unicode.IsDigit(rune(l.ch)) {
			sl++
			l.read()
		}
		s := l.buf[p : p+sl]
		return Token{kind: TokenKindInt, val: string(s)}
	case unicode.IsLetter(rune(l.ch)) || l.ch == '_':
		p := l.pos
		sl := 0
		for unicode.IsLetter(rune(l.ch)) || l.ch == '_' {
			sl++
			l.read()
		}
		s := string(l.buf[p : p+sl])
		switch {
		case s == "input":
			return Token{kind: TokenKindInput}
		case s == "output":
			return Token{kind: TokenKindOutput}
		case s == "goto":
			return Token{kind: TokenKindGoto}
		case s == "if":
			return Token{kind: TokenKindIf}
		case s == "then":
			return Token{kind: TokenKindThen}
		default:
			return Token{kind: TokenKindIdent, val: s}
		}
	default:
		s := l.buf[l.pos : l.pos+1]
		l.read()
		return Token{kind: TokenKindInvalid, val: string(s)}
	}
}

func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(rune(l.ch)) {
		l.read()
	}
}

func createLexer(buf []byte) Lexer {
	var l Lexer
	l.buf = buf
	l.read()
	return l
}

func tokenize(buf []byte) []Token {
	l := createLexer(buf)
	var ts []Token
	var t Token
	for ok := true; ok; ok = t.kind != TokenKindEnd {
		t = l.nextToken()
		ts = append(ts, t)
	}
	return ts
}
