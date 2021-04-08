package parser

type NodeType string

var (
	ElementNode NodeType = "Element"
	RootNode    NodeType = "Root"
	TextNode    NodeType = "Text"
	UnknownNode NodeType = "Unknown"
)

type Node struct {
	Name       string
	Type       NodeType
	Children   []Node
	Attributes []Attribute
}

type Attribute struct {
	Name  string
	Value string
}
