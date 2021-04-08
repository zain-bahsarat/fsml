package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripEndTag(t *testing.T) {
	assert.Equal(t, "EndTag", stripEndTag("</EndTag>"))
}

func TestStripBeginTag(t *testing.T) {
	assert.Equal(t, "BeginTag", stripBeginTag("<BeginTag>"))
}
