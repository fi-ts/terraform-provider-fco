package vlans

import (
	mvmv1 "github.com/fi-ts/api/go/fits/api/mvm/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// stringOrNull maps a nullable *string proto field to a Terraform string,
// yielding a null value for a nil pointer.
func stringOrNull(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

// vlanViewFromAPI maps an API VLAN to its read-only view.
func vlanViewFromAPI(v *mvmv1.VLAN) vlanView {
	return vlanView{
		UUID:         types.StringValue(v.GetUuid()),
		VlanID:       types.Int64Value(int64(v.GetVlanId())),
		SubnetTitle:  types.StringValue(v.GetSubnetTitle()),
		SubnetCidr:   types.StringValue(v.GetSubnetCidr()),
		LocationUUID: types.StringValue(v.GetLocationUuid()),
		Tenant:       types.StringValue(v.GetTenant()),
		DomainUUID:   stringOrNull(v.DomainUuid),
	}
}
