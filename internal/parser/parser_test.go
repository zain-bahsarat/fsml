package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {

	testcases := []struct {
		input    string
		expected Node
	}{
		{
			input: `<Schema>
			<OnBefore>
				<Task>task1</Task>
			</OnBefore>
				<States>
					<new>
						<OnBefore>
							<Task>task1</Task>
						</OnBefore>
						<Events>
							<DummyEvent targetState="pending" errorState="error">
								<Task>t1</Task>
								<Task>t2</Task>
							</DummyEvent>
						</Events>
					</new>
				</States>
			</Schema>`,
			expected: Node{
				Name: "Schema",
				Type: RootNode,
				Children: []Node{
					{
						Name: "OnBefore",
						Type: ElementNode,
						Children: []Node{{
							Name: "Task",
							Type: ElementNode,
							Children: []Node{{
								Name: "task1",
								Type: TextNode,
							}},
						}},
					},
					{
						Name: "States",
						Type: ElementNode,
						Children: []Node{{
							Name: "new",
							Type: ElementNode,
							Children: []Node{
								{
									Name: "OnBefore",
									Type: ElementNode,
									Children: []Node{{
										Name: "Task",
										Type: ElementNode,
										Children: []Node{{
											Name: "task1",
											Type: TextNode,
										}},
									}},
								}, {
									Name: "Events",
									Type: ElementNode,
									Children: []Node{{
										Name: "DummyEvent",
										Type: ElementNode,
										Attributes: []Attribute{
											{Name: "targetState", Value: "pending"},
											{Name: "errorState", Value: "error"},
										},
										Children: []Node{{
											Name: "Task",
											Type: ElementNode,
											Children: []Node{{
												Name: "t1",
												Type: TextNode,
											}},
										},
											{
												Name: "Task",
												Type: ElementNode,
												Children: []Node{{
													Name: "t2",
													Type: TextNode,
												}},
											}},
									}},
								}},
						}},
					}},
				Attributes: make([]Attribute, 0),
			},
		},
	}

	for _, tt := range testcases {

		parser := New(NewLexer(tt.input))

		tree := parser.Parse()
		assert.Equal(t, tt.expected.Name, tree.Name)
		assert.Equal(t, tt.expected.Type, tree.Type)
		assert.True(t, len(tree.Children) > 0, "Wrong children count")

		assert.Equal(t, tt.expected.Children, tree.Children, "childrens are not equal")
	}
}
