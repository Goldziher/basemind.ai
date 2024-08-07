package ptr_test

import (
	"github.com/basemind-ai/monorepo/internal/utils/ptr"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTo(t *testing.T) {
	type T int

	val := T(0)
	pointer := ptr.To(val)
	assert.Equal(t, val, *pointer)

	val = T(1)
	pointer = ptr.To(val)
	assert.Equal(t, val, *pointer)
}

func TestDeref(t *testing.T) {
	type T int

	var val, def T = 1, 0

	out := ptr.Deref(&val, def)
	assert.Equal(t, val, out)

	out = ptr.Deref(nil, def)
	assert.Equal(t, def, out)
}
