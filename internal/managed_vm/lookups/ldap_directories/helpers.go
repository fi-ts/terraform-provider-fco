package ldapdirectories

import (
	mvmv1 "github.com/fi-ts/api/go/fits/api/mvm/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ldapDirectoryViewFromAPI maps an API LDAP directory to its read-only view.
func ldapDirectoryViewFromAPI(l *mvmv1.LDAP) ldapDirectoryView {
	return ldapDirectoryView{
		UUID:   types.StringValue(l.GetUuid()),
		Fqdn:   types.StringValue(l.GetFqdn()),
		Tenant: types.StringValue(l.GetTenant()),
	}
}
