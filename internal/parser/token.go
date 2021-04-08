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
	for _, t := range []string{">", "<"} {
		tag = strings.ReplaceAll(tag, t, "")
	}
	return tag
}

func stripEndTag(tag string) string {
	for _, t := range []string{">", "<", "/"} {
		tag = strings.ReplaceAll(tag, t, "")
	}
	return tag
}
