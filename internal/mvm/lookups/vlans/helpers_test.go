package vlans

import (
	"testing"

	mvmv1 "github.com/fi-ts/api/go/fits/api/mvm/v1"
	"github.com/stretchr/testify/assert"
)

func TestVlanViewFromAPI(t *testing.T) {
	v := vlanViewFromAPI(&mvmv1.VLAN{
		Uuid:         "vlan-uuid",
		VlanId:       2001,
		SubnetTitle:  "pod01-vlan2001 (Production)",
		SubnetCidr:   "10.0.1.0/24",
		LocationUuid: "loc-uuid",
		Tenant:       "tenant-abc",
		// DomainUuid intentionally nil -> domain_uuid is null.
	})
	assert.Equal(t, "vlan-uuid", v.UUID.ValueString())
	assert.EqualValues(t, 2001, v.VlanID.ValueInt64())
	assert.Equal(t, "pod01-vlan2001 (Production)", v.SubnetTitle.ValueString())
	assert.Equal(t, "10.0.1.0/24", v.SubnetCidr.ValueString())
	assert.Equal(t, "loc-uuid", v.LocationUUID.ValueString())
	assert.Equal(t, "tenant-abc", v.Tenant.ValueString())
	assert.True(t, v.DomainUUID.IsNull())
}

func TestVlanViewFromAPIDomainSet(t *testing.T) {
	dom := "dom-uuid"
	v := vlanViewFromAPI(&mvmv1.VLAN{Uuid: "vlan-uuid", DomainUuid: &dom})
	assert.Equal(t, "dom-uuid", v.DomainUUID.ValueString())
	assert.False(t, v.DomainUUID.IsNull())
}

func TestStringOrNull(t *testing.T) {
	assert.True(t, stringOrNull(nil).IsNull())
	s := "dom.example"
	assert.Equal(t, "dom.example", stringOrNull(&s).ValueString())
}
