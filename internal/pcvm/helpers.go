package pcvm

import (
	pcvmv1 "github.com/fi-ts/api/go/fits/api/pcvm/v1"
	"github.com/fi-ts/terraform-provider-fco/internal/shared"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// publicCloudVMViewFromAPI maps an API VM to its shared read-only view.
func publicCloudVMViewFromAPI(vm *pcvmv1.PublicCloudVM) publicCloudVMView {
	status := types.ObjectNull(statusAttrTypes)
	if vm.Status != nil {
		status, _ = types.ObjectValue(statusAttrTypes, map[string]attr.Value{
			"state":     types.StringValue(stateToString(vm.Status.State)),
			"operation": types.StringValue(operationToString(vm.Status.Operation)),
			"info":      types.StringValue(vm.Status.Info),
		})
	}
	return publicCloudVMView{
		UUID:   types.StringValue(vm.Uuid),
		Name:   types.StringValue(vm.Name),
		Size:   types.StringValue(sizeToString(vm.Size)),
		Labels: shared.LabelsFromAPI(vm.GetMeta().GetLabels()),
		Status: status,
	}
}

// publicCloudVMFromAPI maps an API VM to the resource/data source model.
func publicCloudVMFromAPI(vm *pcvmv1.PublicCloudVM) publicCloudVMModel {
	return publicCloudVMModel{
		publicCloudVMView: publicCloudVMViewFromAPI(vm),
		Tenant:            types.StringValue(vm.Tenant),
		ProjectUUID:       types.StringValue(vm.ProjectUuid),
		UpdatedAt:         types.StringValue(shared.TimestampFromAPI(vm.GetMeta().GetUpdatedAt())),
	}
}
