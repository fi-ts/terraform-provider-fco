package pcvm

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// publicCloudVMView holds the attributes every representation of a public
// cloud VM shares: the items of the fco_public_cloud_vms list data source and
// the core of publicCloudVMModel.
type publicCloudVMView struct {
	UUID   types.String `tfsdk:"uuid"`
	Name   types.String `tfsdk:"name"`
	Size   types.String `tfsdk:"size"`
	Labels types.Map    `tfsdk:"labels"`
	Status types.Object `tfsdk:"status"`
}

// statusAttrTypes describes the nested status object.
var statusAttrTypes = map[string]attr.Type{
	"state":     types.StringType,
	"operation": types.StringType,
	"info":      types.StringType,
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
	ID          types.String        `tfsdk:"id"`
	ProjectUUID types.String        `tfsdk:"project_uuid"`
	Name        types.String        `tfsdk:"name"`
	Labels      types.Map           `tfsdk:"labels"`
	VMs         []publicCloudVMView `tfsdk:"vms"`
}
