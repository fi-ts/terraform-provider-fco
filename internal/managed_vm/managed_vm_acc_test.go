package managedvm_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/fi-ts/terraform-provider-fco/internal/provider"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"fco": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// testAccPreCheck requires the FCO_* credentials to be exported.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	for _, key := range []string{"FCO_TOKEN", "FCO_BASE_URL", "FCO_PROJECT_UUID", "FCO_TENANT"} {
		if os.Getenv(key) == "" {
			t.Fatalf("environment variable %s must be set for acceptance tests", key)
		}
	}
}

// TestAccManagedVMLookups exercises the read-only list data sources backed by the
// OS, Location, Domain, LDAP and VLAN services. Gated on TF_ACC.
func TestAccManagedVMLookups(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccManagedVMLookupsConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.fco_managed_vm_operating_systems.all", "id"),
					resource.TestCheckResourceAttrSet("data.fco_managed_vm_locations.all", "id"),
					resource.TestCheckResourceAttrSet("data.fco_managed_vm_locations.all", "locations.#"),
					resource.TestCheckResourceAttrSet("data.fco_managed_vm_domains.in_location", "id"),
					resource.TestCheckResourceAttrSet("data.fco_managed_vm_ldap_directories.all", "id"),
					resource.TestCheckResourceAttrSet("data.fco_managed_vm_vlans.in_location", "id"),
				),
			},
		},
	})
}

const testAccManagedVMLookupsConfig = `
data "fco_managed_vm_operating_systems" "all" {
}

data "fco_managed_vm_locations" "all" {
}

# Domain and VLAN listings require a datacenter location (the API rejects an
# empty location_uuid), so feed the first available location into both.
data "fco_managed_vm_domains" "in_location" {
  location_uuid = data.fco_managed_vm_locations.all.locations[0].uuid
}

data "fco_managed_vm_ldap_directories" "all" {
}

data "fco_managed_vm_vlans" "in_location" {
  location_uuid = data.fco_managed_vm_locations.all.locations[0].uuid
}
`
