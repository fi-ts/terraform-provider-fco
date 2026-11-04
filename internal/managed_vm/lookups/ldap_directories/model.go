package ldapdirectories

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ldapDirectoryView holds the read-only attributes of a single LDAP directory
// item in the fco_managed_vm_ldap_directories data source.
type ldapDirectoryView struct {
	UUID   types.String `tfsdk:"uuid"`
	Fqdn   types.String `tfsdk:"fqdn"`
	Tenant types.String `tfsdk:"tenant"`
}

// listLDAPDirectoriesModel describes the Terraform state for the
// fco_managed_vm_ldap_directories data source.
type listLDAPDirectoriesModel struct {
	ID              types.String        `tfsdk:"id"`
	LDAPDirectories []ldapDirectoryView `tfsdk:"ldap_directories"`
}
