package shared

import (
	"context"

	apiv1 "github.com/fi-ts/api/go/fits/api/v1"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ListFilters converts the optional `name` and `labels` filter attributes of a
// list data source into API request values. name is nil when null or empty;
// labels is nil when null, unknown or empty.
func ListFilters(ctx context.Context, name types.String, labels types.Map) (*string, *apiv1.Labels, diag.Diagnostics) {
	var n *string
	if !name.IsNull() && name.ValueString() != "" {
		v := name.ValueString()
		n = &v
	}

	if labels.IsNull() || labels.IsUnknown() {
		return n, nil, nil
	}
	var m map[string]string
	diags := labels.ElementsAs(ctx, &m, false)
	if diags.HasError() {
		return nil, nil, diags
	}
	if len(m) == 0 {
		return n, nil, diags
	}
	return n, &apiv1.Labels{Labels: m}, diags
}
