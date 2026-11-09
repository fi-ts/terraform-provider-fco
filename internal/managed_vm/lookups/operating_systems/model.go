package operatingsystems

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// operatingSystemView holds the read-only attributes of a single operating
// system item in the fco_managed_vm_operating_systems data source.
type operatingSystemView struct {
	UUID  types.String `tfsdk:"uuid"`
	Title types.String `tfsdk:"title"`
	Type  types.String `tfsdk:"type"`
}

// listOperatingSystemsModel describes the Terraform state for the
// fco_managed_vm_operating_systems data source.
type listOperatingSystemsModel struct {
	ID               types.String          `tfsdk:"id"`
	OperatingSystems []operatingSystemView `tfsdk:"operating_systems"`
}
