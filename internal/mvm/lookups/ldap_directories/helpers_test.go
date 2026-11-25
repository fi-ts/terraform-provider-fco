package ldapdirectories

import (
	"testing"

	mvmv1 "github.com/fi-ts/api/go/fits/api/mvm/v1"
	"github.com/stretchr/testify/assert"
)

func TestLDAPDirectoryViewFromAPI(t *testing.T) {
	v := ldapDirectoryViewFromAPI(&mvmv1.LDAP{Uuid: "ldap-uuid", Fqdn: "ldap.example", Tenant: "tenant-abc"})
	assert.Equal(t, "ldap-uuid", v.UUID.ValueString())
	assert.Equal(t, "ldap.example", v.Fqdn.ValueString())
	assert.Equal(t, "tenant-abc", v.Tenant.ValueString())
}
