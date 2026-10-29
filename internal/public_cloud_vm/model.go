package publiccloudvm

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// publicCloudVMView holds the attributes every representation of a public
// cloud VM shares: the items of the fco_public_cloud_vms list data source and
// the core of publicCloudVMModel.
type publicCloudVMView struct {
	UUID            types.String `tfsdk:"uuid"`
	Name            types.String `tfsdk:"name"`
	Size            types.String `tfsdk:"size"`
	Labels          types.Map    `tfsdk:"labels"`
	StatusState     types.String `tfsdk:"status_state"`
	StatusOperation types.String `tfsdk:"status_operation"`
	StatusInfo      types.String `tfsdk:"status_info"`
}

// publicCloudVMModel describes the Terraform state for the fco_public_cloud_vm
// resource and the fco_public_cloud_vm data source.
type publicCloudVMModel struct {
	publicCloudVMView

	Tenant      types.String `tfsdk:"tenant"`
	ProjectUUID types.String `tfsdk:"project_uuid"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// listPublicCloudVMModel describes the Terraform state for the
// fco_public_cloud_vms data source.
type listPublicCloudVMModel struct {
	ID     types.String        `tfsdk:"id"`
	Name   types.String        `tfsdk:"name"`
	Labels types.Map           `tfsdk:"labels"`
	VMs    []publicCloudVMView `tfsdk:"vms"`
}
