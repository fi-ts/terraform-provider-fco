package shared

import (
	"context"

	apiv1 "github.com/fi-ts/api/go/fits/api/v1"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// LabelsToAPI converts a Terraform labels map (or null/unknown) to the API
// Labels message. A null/unknown/empty map yields a nil Labels (no labels are
// set on the entity).
func LabelsToAPI(m types.Map) *apiv1.Labels {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	var elems map[string]string
	if diags := m.ElementsAs(context.Background(), &elems, false); diags.HasError() || len(elems) == 0 {
		return nil
	}
	return &apiv1.Labels{Labels: elems}
}

// LabelsFromAPI converts an API Labels message (possibly nil) to a Terraform
// map. A nil/empty message maps to a null map.
func LabelsFromAPI(l *apiv1.Labels) types.Map {
	if l == nil || len(l.Labels) == 0 {
		return types.MapNull(types.StringType)
	}
	elems := make(map[string]attr.Value, len(l.Labels))
	for k, v := range l.Labels {
		elems[k] = types.StringValue(v)
	}
	m, _ := types.MapValue(types.StringType, elems)
	return m
}

// labelsToMap returns the string map held by m, or an empty map when m is null
// or unknown. It swallows element errors, returning the empty map, so callers
// can rely on a usable map for diffing.
func labelsToMap(m types.Map) map[string]string {
	out := map[string]string{}
	if m.IsNull() || m.IsUnknown() {
		return out
	}
	if diags := m.ElementsAs(context.Background(), &out, false); diags.HasError() {
		return map[string]string{}
	}
	return out
}

// LabelsDelta returns the update/remove diff needed to make stateLabels match
// planLabels (null/unknown count as empty), or nil when nothing changes.
func LabelsDelta(stateLabels, planLabels types.Map) *apiv1.UpdateLabels {
	current := labelsToMap(stateLabels)
	desired := labelsToMap(planLabels)

	update := map[string]string{}
	var remove []string
	for k, v := range desired {
		if cur, ok := current[k]; !ok || cur != v {
			update[k] = v
		}
	}
	for k := range current {
		if _, ok := desired[k]; !ok {
			remove = append(remove, k)
		}
	}

	if len(update) == 0 && len(remove) == 0 {
		return nil
	}

	res := &apiv1.UpdateLabels{}
	if len(update) > 0 {
		res.Update = &apiv1.Labels{Labels: update}
	}
	if len(remove) > 0 {
		res.Remove = remove
	}
	return res
}
