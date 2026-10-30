package publiccloudvm

import (
	"context"
	"fmt"
	"time"

	pcvmv1 "github.com/fi-ts/api/go/fits/api/pcvm/v1"
	apiv1 "github.com/fi-ts/api/go/fits/api/v1"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Terraform attribute values for the size attribute.
const (
	SizeSmall  = "small"
	SizeMedium = "medium"
	SizeLarge  = "large"
)

// sizeUnspecified is reported for sizes the API does not classify as
// small/medium/large.
const sizeUnspecified = "unspecified"

// sizeByString maps the accepted Terraform size strings to the API enum.
// It is the single source of truth for the allowed values.
var sizeByString = map[string]pcvmv1.Size{
	SizeSmall:  pcvmv1.Size_SIZE_SMALL,
	SizeMedium: pcvmv1.Size_SIZE_MEDIUM,
	SizeLarge:  pcvmv1.Size_SIZE_LARGE,
}

// sizeRank maps a size to its ordinal rank (higher = larger), used to
// enforce that only upgrades are allowed. Unknown sizes rank 0 (smallest).
var sizeRank = map[pcvmv1.Size]int{
	pcvmv1.Size_SIZE_SMALL:  1,
	pcvmv1.Size_SIZE_MEDIUM: 2,
	pcvmv1.Size_SIZE_LARGE:  3,
}

// validSizes is the set of VM sizes accepted by the API.
var validSizes = []string{SizeSmall, SizeMedium, SizeLarge}

// sizeValidator restricts the size attribute to the allowed enum values.
var sizeValidator = []validator.String{
	stringvalidator.OneOf(validSizes...),
}

// sizeFromString maps a Terraform size string to the API enum.
// It returns ok=false for unknown values.
func sizeFromString(s string) (pcvmv1.Size, bool) {
	size, ok := sizeByString[s]
	return size, ok
}

// sizeToString maps the API size enum to a Terraform size string.
func sizeToString(s pcvmv1.Size) string {
	for str, v := range sizeByString {
		if v == s {
			return str
		}
	}
	return sizeUnspecified
}

// isSizeDowngrade reports whether newSize is strictly smaller than curSize.
func isSizeDowngrade(curSize, newSize pcvmv1.Size) bool {
	return sizeRank[newSize] < sizeRank[curSize]
}

// stateToString maps the API health-state enum to a Terraform string.
func stateToString(s pcvmv1.State) string {
	switch s {
	case pcvmv1.State_STATE_HEALTHY:
		return "healthy"
	case pcvmv1.State_STATE_WARNING:
		return "warning"
	case pcvmv1.State_STATE_ERROR:
		return "error"
	default:
		return "unspecified"
	}
}

// operationToString maps the API operation enum to a Terraform string.
func operationToString(op pcvmv1.Operation) string {
	switch op {
	case pcvmv1.Operation_OPERATION_CREATING:
		return "creating"
	case pcvmv1.Operation_OPERATION_ACTIVE:
		return "active"
	case pcvmv1.Operation_OPERATION_UPDATING:
		return "updating"
	case pcvmv1.Operation_OPERATION_DELETED:
		return "deleted"
	case pcvmv1.Operation_OPERATION_ERROR:
		return "error"
	default:
		return "unspecified"
	}
}

// labelsToAPI converts a Terraform labels map (or null/unknown) to the API Labels message.
// A null/unknown/empty map yields a nil Labels (no labels are set on the entity).
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

// labelsFromAPI converts an API Labels message (possibly nil) to a Terraform map.
// A nil/empty message maps to a null map.
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

// labelsDelta returns the update/remove diff needed to make stateLabels
// match planLabels (null/unknown count as empty), or nil when nothing
// changes.
func labelsDelta(stateLabels, planLabels types.Map) *apiv1.UpdateLabels {
	current := labelsToMap(stateLabels)
	desired := labelsToMap(planLabels)

	update := map[string]string{}
	remove := []string{}
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

// updatedAtFromAPI formats the API meta.updated_at timestamp as RFC 3339 Nano.
func updatedAtFromAPI(meta *apiv1.Meta) string {
	if meta == nil || meta.UpdatedAt == nil {
		return ""
	}
	return meta.UpdatedAt.AsTime().Format(time.RFC3339Nano)
}

// parseUpdatedAt parses an RFC 3339 / RFC 3339 Nano timestamp back into a protobuf Timestamp.
func parseUpdatedAt(s string) (*timestamppb.Timestamp, error) {
	if s == "" {
		return nil, fmt.Errorf("no updated_at timestamp available for optimistic locking")
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at %q: %w", s, err)
	}
	return timestamppb.New(t), nil
}

// publicCloudVMViewFromAPI maps an API VM to its shared read-only view.
func publicCloudVMViewFromAPI(vm *pcvmv1.PublicCloudVM) publicCloudVMView {
	var statusState, statusOp, statusInfo string
	if vm.Status != nil {
		statusState = stateToString(vm.Status.State)
		statusOp = operationToString(vm.Status.Operation)
		statusInfo = vm.Status.Info
	}
	return publicCloudVMView{
		UUID:            types.StringValue(vm.Uuid),
		Name:            types.StringValue(vm.Name),
		Size:            types.StringValue(sizeToString(vm.Size)),
		Labels:          labelsFromAPI(vm.GetMeta().GetLabels()),
		StatusState:     types.StringValue(statusState),
		StatusOperation: types.StringValue(statusOp),
		StatusInfo:      types.StringValue(statusInfo),
	}
}

// publicCloudVMFromAPI maps an API VM to the resource/data source model.
func publicCloudVMFromAPI(vm *pcvmv1.PublicCloudVM) publicCloudVMModel {
	return publicCloudVMModel{
		publicCloudVMView: publicCloudVMViewFromAPI(vm),
		Tenant:            types.StringValue(vm.Tenant),
		ProjectUUID:       types.StringValue(vm.ProjectUuid),
		UpdatedAt:         types.StringValue(updatedAtFromAPI(vm.GetMeta())),
	}
}
