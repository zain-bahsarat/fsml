package parser

import (
	"testing"
)

func TestNextToken(t *testing.T) {

	input := `<Schema>
	<OnBefore>
		<Task>task1</Task>
	</OnBefore>
	<OnAfter>
		<Task>task1</Task>
	</OnAfter>
	<OnEnter>
		<Task></Task>
	</OnEnter>
	<OnLeave>
		<Task></Task>
		<Task></Task>
	</OnLeave>
		<States>
			<new>
				<OnLeave>
					<Task>task_name</Task>
					<Task>task_name2</Task>
				</OnLeave>
				<Events>
					<DummyEvent targetState="pending" errorState="error">
						<Task>t1</Task>
					</DummyEvent>
				</Events>
			</new>
		</States>
	</Schema>`

	tests := []struct {
		expected        TokenType
		expectedLiteral string
	}{
		{BeginTag, "<Schema"},
		{EndTag, ">"},

		{BeginTag, "<OnBefore"},
		{EndTag, ">"},
		{BeginTag, "<Task"},
		{EndTag, ">"},
		{String, "task1"},
		{CloseTag, "</Task>"},
		{CloseTag, "</OnBefore>"},

		{BeginTag, "<OnAfter"},
		{EndTag, ">"},
		{BeginTag, "<Task"},
		{EndTag, ">"},
		{String, "task1"},
		{CloseTag, "</Task>"},
		{CloseTag, "</OnAfter>"},

		{BeginTag, "<OnEnter"},
		{EndTag, ">"},
		{BeginTag, "<Task"},
		{EndTag, ">"},
		{CloseTag, "</Task>"},
		{CloseTag, "</OnEnter>"},

		{BeginTag, "<OnLeave"},
		{EndTag, ">"},
		{BeginTag, "<Task"},
		{EndTag, ">"},
		{CloseTag, "</Task>"},
		{BeginTag, "<Task"},
		{EndTag, ">"},
		{CloseTag, "</Task>"},
		{CloseTag, "</OnLeave>"},

		{BeginTag, "<States"},
		{EndTag, ">"},

		{BeginTag, "<new"},
		{EndTag, ">"},

		{BeginTag, "<OnLeave"},
		{EndTag, ">"},
		{BeginTag, "<Task"},
		{EndTag, ">"},
		{String, "task_name"},
		{CloseTag, "</Task>"},
		{BeginTag, "<Task"},
		{EndTag, ">"},
		{String, "task_name2"},
		{CloseTag, "</Task>"},
		{CloseTag, "</OnLeave>"},

		{BeginTag, "<Events"},
		{EndTag, ">"},
		{BeginTag, "<DummyEvent"},
		{String, "targetState"},
		{Assign, "="},
		{DoubleQuote, "\""},
		{String, "pending"},
		{DoubleQuote, "\""},
		{String, "errorState"},
		{Assign, "="},
		{DoubleQuote, "\""},
		{String, "error"},
		{DoubleQuote, "\""},
		{EndTag, ">"},
		{BeginTag, "<Task"},
		{EndTag, ">"},
		{String, "t1"},
		{CloseTag, "</Task>"},
		{CloseTag, "</DummyEvent>"},
		{CloseTag, "</Events>"},
		{CloseTag, "</new>"},
		{CloseTag, "</States>"},
		{CloseTag, "</Schema>"},
	}

	lex := NewLexer(input)

	for i, tt := range tests {

		tok := lex.NextToken()
		if tok.Type != tt.expected {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q", i, tt.expected, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - token literal wrong. expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}
