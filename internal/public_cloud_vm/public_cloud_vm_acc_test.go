package publiccloudvm_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

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

// TestAccPublicCloudVM exercises the full CRUD + import lifecycle of the
// fco_public_cloud_vm resource. Gated on TF_ACC.
func TestAccPublicCloudVM(t *testing.T) {
	name := acctest.RandomWithPrefix("fco-pcvm")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create.
			{
				Config: testAccConfig(name, "small", nil),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fco_public_cloud_vm.test", "name", name),
					resource.TestCheckResourceAttr("fco_public_cloud_vm.test", "size", "small"),
					resource.TestCheckResourceAttrSet("fco_public_cloud_vm.test", "uuid"),
					resource.TestCheckResourceAttr("fco_public_cloud_vm.test", "status_operation", "active"),
				),
			},
			// Import + re-read — verifies Read is idempotent and import works.
			{
				ResourceName:      "fco_public_cloud_vm.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The resource has no `id` attribute (it is keyed by `uuid`),
				// so derive the import id from prior state.
				ImportStateIdFunc: testAccPublicCloudVMImportID,
			},
			// Upgrade (size small -> medium).
			{
				Config: testAccConfig(name, "medium", map[string]string{"env": "test"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fco_public_cloud_vm.test", "size", "medium"),
					resource.TestCheckResourceAttr("fco_public_cloud_vm.test", "labels.env", "test"),
				),
			},
			// Rename.
			{
				Config: testAccConfig(name+"-renamed", "medium", map[string]string{"env": "test"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fco_public_cloud_vm.test", "name", name+"-renamed"),
				),
			},
			// Label update: overwrite `env` and add `owner` — verifies label
			// diffs are reconciled on update.
			{
				Config: testAccConfig(name+"-renamed", "medium", map[string]string{"env": "prod", "owner": "fco"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fco_public_cloud_vm.test", "labels.env", "prod"),
					resource.TestCheckResourceAttr("fco_public_cloud_vm.test", "labels.owner", "fco"),
				),
			},
			// Data sources: singular (by uuid) and list (by name).
			{
				Config: testAccConfig(name+"-renamed", "medium", map[string]string{"env": "test"}) +
					testAccDataSources(name+"-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.fco_public_cloud_vm.by_uuid", "uuid"),
					resource.TestCheckResourceAttr("data.fco_public_cloud_vm.by_uuid", "name", name+"-renamed"),
					resource.TestCheckResourceAttr("data.fco_public_cloud_vms.by_name", "vms.#", "1"),
					resource.TestCheckResourceAttr("data.fco_public_cloud_vms.by_name", "vms.0.name", name+"-renamed"),
				),
			},
		},
	})
}

// testAccConfig builds the resource HCL.
func testAccConfig(name, size string, labels map[string]string) string {
	return testAccProviderConfig() +
		`
resource "fco_public_cloud_vm" "test" {
  name = "` + name + `"
  size = "` + size + `"
` + labelsHCL(labels) + `
}
`
}

// testAccDataSources builds the singular and list data-source HCL.
func testAccDataSources(name string) string {
	return `
data "fco_public_cloud_vm" "by_uuid" {
  uuid = fco_public_cloud_vm.test.uuid
}

data "fco_public_cloud_vms" "by_name" {
  name = "` + name + `"
}
`
}

func labelsHCL(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	hcl := "  labels = {\n"
	for k, v := range labels {
		hcl += "    " + k + ` = "` + v + `"\n`
	}
	hcl += "  }\n"
	return hcl
}

// testAccProviderConfig is empty: the provider reads credentials from the
// FCO_* environment variables.
func testAccProviderConfig() string {
	return ""
}

// testAccPublicCloudVMImportID reads the `uuid` from prior state and uses it
// as the import identifier; the resource has no `id` attribute.
func testAccPublicCloudVMImportID(s *terraform.State) (string, error) {
	for _, m := range s.Modules {
		if r, ok := m.Resources["fco_public_cloud_vm.test"]; ok && r.Primary != nil {
			if uuid, ok := r.Primary.Attributes["uuid"]; ok {
				return uuid, nil
			}
		}
	}
	return "", fmt.Errorf("resource fco_public_cloud_vm.test not found in prior state")
}
