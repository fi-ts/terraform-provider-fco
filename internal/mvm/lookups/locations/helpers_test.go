package locations

import (
	"testing"

	mvmv1 "github.com/fi-ts/api/go/fits/api/mvm/v1"
	"github.com/stretchr/testify/assert"
)

func TestLocationViewFromAPI(t *testing.T) {
	v := locationViewFromAPI(&mvmv1.Location{Uuid: "loc-uuid", Title: "Nürnberg (RZ1)"})
	assert.Equal(t, "loc-uuid", v.UUID.ValueString())
	assert.Equal(t, "Nürnberg (RZ1)", v.Title.ValueString())
}
