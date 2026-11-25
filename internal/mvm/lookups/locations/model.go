package locations

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// locationView holds the read-only attributes of a single location item in
// the fco_managed_vm_locations data source.
type locationView struct {
	UUID  types.String `tfsdk:"uuid"`
	Title types.String `tfsdk:"title"`
}

// listLocationsModel describes the Terraform state for the fco_managed_vm_locations
// data source.
type listLocationsModel struct {
	ID        types.String   `tfsdk:"id"`
	Locations []locationView `tfsdk:"locations"`
}
