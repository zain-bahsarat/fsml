package schema

import (
	"fmt"
	"testing"

	"github.com/zain-bahsarat/fsml/internal/parser"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	testcases := []struct {
		input    string
		expected *Schema
	}{
		{
			input: `<Schema>
			<OnBeforeEvent>
				<Task>task1</Task>
			</OnBeforeEvent>
				<States>
					<new>
						<OnBeforeEvent>
							<Task>task1</Task>
						</OnBeforeEvent>
						<Events>
							<DummyEvent targetState="pending" errorState="error">
								<Task>t1</Task>
								<Task>t2</Task>
							</DummyEvent>
						</Events>
					</new>
				</States>
			</Schema>`,
			expected: &Schema{
				DefaultEvents: DefaultEvents{OnBeforeEvent: Event{Tasks: []string{"task1"}}},
				States:        []State{{Name: "new", DefaultEvents: DefaultEvents{OnBeforeEvent: Event{Tasks: []string{"task1"}}}, Events: []CustomEvent{{Name: "DummyEvent", TargetState: "pending", ErrorState: "error", Tasks: []string{"t1", "t2"}}}}},
			},
		},
	}

	for i, tt := range testcases {

		p := parser.New(parser.NewLexer(tt.input))
		s, err := New(p)

		assert.Nil(t, err)
		assert.Equal(t, tt.expected, s, fmt.Sprintf("tests[%d] - schema error", i))
	}
}
