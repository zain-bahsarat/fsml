package parser

import "strings"

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	//Special
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Delimiters
	BeginTag = "BeginTag" // <tag
	EndTag   = "EndTag"   // >
	CloseTag = "CloseTag" // </tag>

	Assign      = "="
	DoubleQuote = "\""

	// Types
	String = "String"
)

func stripBeginTag(tag string) string {
	return strings.Map(func(r rune) rune {
		if r == '>' || r == '<' {
			return -1 // remove character
		}
		return r
	}, tag)
}

func stripEndTag(tag string) string {
	return strings.Map(func(r rune) rune {
		if r == '>' || r == '<' || r == '/' {
			return -1 // remove character
		}
		return r
	}, tag)
}
