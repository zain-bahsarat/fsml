package parser

type Lexer struct {
	input        string
	position     int // current caracter position
	readPosition int //(next character in input)
	ch           byte
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0x00
	} else {
		l.ch = l.input[l.readPosition]
	}

	l.position = l.readPosition
	if l.ch != 0x00 {
		l.readPosition += 1
	}
}

func (l *Lexer) NextToken() Token {
	var tok Token

	// skip whitespace characters
	l.skipWhitespace()

	switch l.ch {
	case '<':
		if l.peekChar() == '/' {
			tok.Type = CloseTag
			tok.Literal = l.readCloseTag()
			return tok
		} else if isLetter(l.peekChar()) {
			tok.Literal = l.readStartTag()
			tok.Type = BeginTag
			return tok
		} else {
			tok = newToken(ILLEGAL, l.ch)
		}
	case '>':
		tok = newToken(EndTag, l.ch)
	case '"':
		tok = newToken(DoubleQuote, l.ch)
	case '=':
		tok = newToken(Assign, l.ch)
	case 0x00:
		tok.Literal = ""
		tok.Type = EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readString()
			tok.Type = String
			return tok
		} else {
			tok = newToken(ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readStartTag() string {

	s := string(l.ch)
	l.readChar()
	return s + l.readString()
}

func (l *Lexer) readCloseTag() string {
	pos := l.position

	l.readChar()   // one for /
	l.readChar()   // one to reach alpha char
	l.readString() // tagname

	for l.ch != '>' && l.peekChar() != 0x00 {
		l.readChar()
		l.skipWhitespace()
	}

	l.readChar() // ending tag >

	return l.input[pos:l.position]
}

func (l *Lexer) readString() string {
	pos := l.position
	for isLetter(l.ch) {
		l.readChar()
	}

	return l.input[pos:l.position]
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}

	return l.input[l.readPosition]
}

func (l *Lexer) skipWhitespace() {
	for isWhiteSpace(l.ch) {
		l.readChar()
	}
}

func isWhiteSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isLetter(ch byte) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch == '_' || ch >= '0' && ch <= '9'
}

func newToken(tokenType TokenType, ch byte) Token {
	return Token{Type: tokenType, Literal: string(ch)}
}
