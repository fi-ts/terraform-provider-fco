package operatingsystems

import (
	mvmv1 "github.com/fi-ts/api/go/fits/api/mvm/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// osTypeToString maps the API OS type enum to a Terraform string.
func osTypeToString(t mvmv1.OSType) string {
	switch t {
	case mvmv1.OSType_OS_TYPE_LINUX:
		return "linux"
	case mvmv1.OSType_OS_TYPE_WINDOWS:
		return "windows"
	default:
		return "unspecified"
	}
}

// operatingSystemViewFromAPI maps an API OS to its read-only view.
func operatingSystemViewFromAPI(os *mvmv1.OS) operatingSystemView {
	return operatingSystemView{
		UUID:  types.StringValue(os.GetUuid()),
		Title: types.StringValue(os.GetTitle()),
		Type:  types.StringValue(osTypeToString(os.GetType())),
	}
}
