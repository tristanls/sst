package sst

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThing(t *testing.T) {
	assert.Equal(t, "Constitutes", associations["part_of"].SemanticType.String())
	assert.Equal(t, -2, int(associations["part_of"].SemanticType))
}

func TestToDocumentKey(t *testing.T) {
	assert.Equal(t, "Number_12345", toDocumentKey("Number 12345"))
}
