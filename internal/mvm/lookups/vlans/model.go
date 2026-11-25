package vlans

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// vlanView holds the read-only attributes of a single VLAN item in the
// fco_managed_vm_vlans data source.
type vlanView struct {
	UUID         types.String `tfsdk:"uuid"`
	VlanID       types.Int64  `tfsdk:"vlan_id"`
	SubnetTitle  types.String `tfsdk:"subnet_title"`
	SubnetCidr   types.String `tfsdk:"subnet_cidr"`
	LocationUUID types.String `tfsdk:"location_uuid"`
	Tenant       types.String `tfsdk:"tenant"`
	DomainUUID   types.String `tfsdk:"domain_uuid"`
}

// listVLANsModel describes the Terraform state for the fco_managed_vm_vlans data
// source.
type listVLANsModel struct {
	ID           types.String `tfsdk:"id"`
	LocationUUID types.String `tfsdk:"location_uuid"`
	DomainUUID   types.String `tfsdk:"domain_uuid"`
	VLANs        []vlanView   `tfsdk:"vlans"`
}
