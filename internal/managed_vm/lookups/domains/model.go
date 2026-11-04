package domains

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// domainView holds the read-only attributes of a single domain item in the
// fco_managed_vm_domains data source.
type domainView struct {
	UUID   types.String `tfsdk:"uuid"`
	Fqdn   types.String `tfsdk:"fqdn"`
	Tenant types.String `tfsdk:"tenant"`
}

// listDomainsModel describes the Terraform state for the fco_managed_vm_domains data
// source.
type listDomainsModel struct {
	ID           types.String `tfsdk:"id"`
	LocationUUID types.String `tfsdk:"location_uuid"`
	Domains      []domainView `tfsdk:"domains"`
}
