package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListID(t *testing.T) {
	// Deterministic and stable for equal inputs.
	assert.Equal(t, ListID(nil), ListID(nil))
	assert.Equal(t, ListID([]string{"a", "b"}), ListID([]string{"a", "b"}))
	// Different inputs yield different hashes.
	assert.NotEqual(t, ListID([]string{"a"}), ListID([]string{"a", "b"}))
	// Always non-empty.
	assert.NotEmpty(t, ListID(nil).ValueString())
	assert.NotEmpty(t, ListID([]string{"x"}).ValueString())
}
