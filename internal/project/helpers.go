package project

import (
	"context"
	"errors"
	"fmt"
	"time"

	apiv1 "github.com/fi-ts/api/go/fits/api/v1"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// labelsToAPI converts a Terraform labels map (or null/unknown) to the API
// Labels message. A null/unknown/empty map yields a nil Labels (no labels are
// set on the entity).
func labelsToAPI(m types.Map) *apiv1.Labels {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	var elems map[string]string
	if diags := m.ElementsAs(context.Background(), &elems, false); diags.HasError() || len(elems) == 0 {
		return nil
	}
	return &apiv1.Labels{Labels: elems}
}

// labelsFromAPI converts an API Labels message (possibly nil) to a Terraform
// map. A nil/empty message maps to a null map.
func labelsFromAPI(l *apiv1.Labels) types.Map {
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

// labelsDelta returns the update/remove diff needed to make stateLabels match
// planLabels (null/unknown count as empty), or nil when nothing changes.
func labelsDelta(stateLabels, planLabels types.Map) *apiv1.UpdateLabels {
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

// timestampFromAPI formats a protobuf Timestamp (possibly nil) as RFC 3339
// Nano, or "" when absent.
func timestampFromAPI(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().Format(time.RFC3339Nano)
}

// errNoUpdatedAt is returned by parseUpdatedAt when the stored value is empty
// (no timestamp available for optimistic locking). Callers can distinguish
// this from a genuine parse error with errors.Is.
var errNoUpdatedAt = errors.New("no updated_at timestamp available for optimistic locking")

// parseUpdatedAt parses an RFC 3339 / RFC 3339 Nano timestamp back into a
// protobuf Timestamp.
func parseUpdatedAt(s string) (*timestamppb.Timestamp, error) {
	if s == "" {
		return nil, errNoUpdatedAt
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at %q: %w", s, err)
	}
	return timestamppb.New(t), nil
}

// avatarURLFromAPI maps the API's optional avatar_url (possibly nil or empty)
// to a Terraform string. An absent/empty value maps to a null string so that
// omitting the attribute in configuration does not produce a spurious diff.
func avatarURLFromAPI(s *string) types.String {
	if s == nil || *s == "" {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

// projectViewFromAPI maps an API project to its shared read-only view.
func projectViewFromAPI(p *apiv1.Project) projectView {
	return projectView{
		UUID:        types.StringValue(p.GetUuid()),
		Name:        types.StringValue(p.GetName()),
		Description: types.StringValue(p.GetDescription()),
		AvatarURL:   avatarURLFromAPI(p.AvatarUrl),
		Labels:      labelsFromAPI(p.GetMeta().GetLabels()),
		Tenant:      types.StringValue(p.GetTenant()),
		CreatedAt:   types.StringValue(timestampFromAPI(p.GetMeta().GetCreatedAt())),
		UpdatedAt:   types.StringValue(timestampFromAPI(p.GetMeta().GetUpdatedAt())),
	}
}

// projectFromAPI maps an API project to the resource/data source model.
func projectFromAPI(p *apiv1.Project) projectModel {
	return projectModel{
		projectView: projectViewFromAPI(p),
	}
}
