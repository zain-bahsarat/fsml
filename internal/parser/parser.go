package parser

import (
	"fmt"
	"strings"
)

type Parser struct {
	l      *Lexer
	errors []string

	curToken  Token
	peekToken Token
}

func New(l *Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) expect(t TokenType) bool {
	if !p.curTokenIs(t) {
		p.peekError(t)
		return false
	}

	return true
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) Parse() *Node {

	node := (*Node)(nil)
	for !p.curTokenIs(EOF) {

		n := p.parseNode()
		if n != nil && node == nil {
			node = n
			node.Type = RootNode
		} else if n != nil {
			node.Children = append(node.Children, *node)
			node = n
		}

		p.nextToken()
	}

	return node
}

func (p *Parser) parseNode() *Node {
	switch p.curToken.Type {
	case BeginTag:
		return p.parseElementNode()
	case String:
		return p.parseTextNode()
	}

	return &Node{}
}

func (p *Parser) parseTextNode() *Node {
	n := Node{Type: TextNode}
	var sb strings.Builder

	sb.WriteString(p.curToken.Literal)
	for !p.peekTokenIs(CloseTag) && !p.peekTokenIs(EOF) {
		sb.WriteString(p.curToken.Literal)
		if !p.expectPeek(String) {
			break
		}
	}

	n.Name = sb.String()
	return &n
}

func (p *Parser) parseElementNode() *Node {

	p.expect(BeginTag)

	n := Node{Type: ElementNode}
	n.Name = stripBeginTag(p.curToken.Literal)
	if attributes := p.parseAttributes(); len(attributes) > 0 {
		n.Attributes = attributes
	}

	p.expectPeek(EndTag) //consume EndTag '>'

	for !p.peekTokenIs(CloseTag) && !p.curTokenIs(EOF) {
		p.nextToken()
		nn := p.parseNode()
		n.Children = append(n.Children, *nn)
	}

	p.expectPeek(CloseTag) //consume CloseTag </tag>
	if n.Name != stripEndTag(p.curToken.Literal) {
		n.Type = UnknownNode
	}

	return &n
}

func (p *Parser) parseAttributes() []Attribute {
	attributes := make([]Attribute, 0)
	for !p.peekTokenIs(EndTag) && !p.peekTokenIs(EOF) {
		attr := p.parseAttribute()
		if attr != nil {
			attributes = append(attributes, *attr)
		} else {
			p.nextToken()
		}
	}

	return attributes
}

func (p *Parser) parseAttribute() *Attribute {
	vals := []string{}
	for _, tok := range []TokenType{String, Assign, DoubleQuote, String, DoubleQuote} {
		if p.expectPeek(tok) {
			if p.curTokenIs(String) {
				vals = append(vals, p.curToken.Literal)
			}
		}
	}

	if len(vals) < 2 {
		return nil
	}

	return &Attribute{Name: vals[0], Value: vals[1]}
}
