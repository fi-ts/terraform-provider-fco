package domains

import (
	"testing"

	mvmv1 "github.com/fi-ts/api/go/fits/api/mvm/v1"
	"github.com/stretchr/testify/assert"
)

func TestDomainViewFromAPI(t *testing.T) {
	v := domainViewFromAPI(&mvmv1.Domain{Uuid: "dom-uuid", Fqdn: "corp.example", Tenant: "tenant-abc"})
	assert.Equal(t, "dom-uuid", v.UUID.ValueString())
	assert.Equal(t, "corp.example", v.Fqdn.ValueString())
	assert.Equal(t, "tenant-abc", v.Tenant.ValueString())
}
