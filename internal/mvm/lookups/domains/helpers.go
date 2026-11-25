package domains

import (
	mvmv1 "github.com/fi-ts/api/go/fits/api/mvm/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// domainViewFromAPI maps an API domain to its read-only view.
func domainViewFromAPI(d *mvmv1.Domain) domainView {
	return domainView{
		UUID:   types.StringValue(d.GetUuid()),
		Fqdn:   types.StringValue(d.GetFqdn()),
		Tenant: types.StringValue(d.GetTenant()),
	}
}
