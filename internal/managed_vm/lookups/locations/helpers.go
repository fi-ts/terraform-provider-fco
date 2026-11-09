package locations

import (
	mvmv1 "github.com/fi-ts/api/go/fits/api/mvm/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// locationViewFromAPI maps an API location to its read-only view.
func locationViewFromAPI(l *mvmv1.Location) locationView {
	return locationView{
		UUID:  types.StringValue(l.GetUuid()),
		Title: types.StringValue(l.GetTitle()),
	}
}
