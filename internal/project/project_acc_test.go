package project_test

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

// testAccPreCheck requires the FCO_* credentials to be exported. Projects are
// tenant scoped.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	for _, key := range []string{"FCO_TOKEN", "FCO_BASE_URL", "FCO_TENANT"} {
		if os.Getenv(key) == "" {
			t.Fatalf("environment variable %s must be set for acceptance tests", key)
		}
	}
}

// TestAccProject exercises the full CRUD + import lifecycle of the
// fco_project resource. Gated on TF_ACC.
func TestAccProject(t *testing.T) {
	name := acctest.RandomWithPrefix("fco-proj")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create.
			{
				Config: testAccConfig(name, "Initial description", nil),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fco_project.test", "name", name),
					resource.TestCheckResourceAttr("fco_project.test", "description", "Initial description"),
					resource.TestCheckResourceAttrSet("fco_project.test", "uuid"),
					resource.TestCheckResourceAttrSet("fco_project.test", "created_at"),
				),
			},
			// Import + re-read — verifies Read is idempotent and import works.
			{
				ResourceName:                         "fco_project.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "uuid",
				ImportStateIdFunc:                    testAccProjectImportID,
			},
			// Update description and labels.
			{
				Config: testAccConfig(name, "Updated description", map[string]string{"env": "test"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fco_project.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("fco_project.test", "labels.env", "test"),
				),
			},
			// Rename. Because `name` is RequiresReplace, this step destroys
			// and recreates the project with a fresh UUID, exercising the
			// replace path end to end.
			{
				Config: testAccConfig(name+"-renamed", "Updated description", map[string]string{"env": "test"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fco_project.test", "name", name+"-renamed"),
				),
			},
			// Data sources: singular (by uuid) and list (by name).
			{
				Config: testAccConfig(name+"-renamed", "Updated description", map[string]string{"env": "test"}) +
					testAccDataSources(name+"-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.fco_project.by_uuid", "uuid"),
					resource.TestCheckResourceAttr("data.fco_project.by_uuid", "name", name+"-renamed"),
					resource.TestCheckResourceAttr("data.fco_projects.by_name", "projects.#", "1"),
					resource.TestCheckResourceAttr("data.fco_projects.by_name", "projects.0.name", name+"-renamed"),
				),
			},
		},
	})
}

// testAccConfig builds the resource HCL.
func testAccConfig(name, description string, labels map[string]string) string {
	return testAccProviderConfig() +
		"\nresource \"fco_project\" \"test\" {\n" +
		"  name = \"" + name + "\"\n" +
		"  description = \"" + description + "\"\n" +
		labelsHCL(labels) +
		"}\n"
}

// testAccDataSources builds the singular and list data-source HCL.
func testAccDataSources(name string) string {
	return "\n" +
		"data \"fco_project\" \"by_uuid\" {\n" +
		"  uuid = fco_project.test.uuid\n" +
		"}\n" +
		"\n" +
		"data \"fco_projects\" \"by_name\" {\n" +
		"  name = \"" + name + "\"\n" +
		"}\n"
}

func labelsHCL(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	hcl := "  labels = {\n"
	for k, v := range labels {
		hcl += "    " + k + " = \"" + v + "\"\n"
	}
	hcl += "  }\n"
	return hcl
}

// testAccProviderConfig is empty: the provider reads credentials from the
// FCO_* environment variables.
func testAccProviderConfig() string {
	return ""
}

// testAccProjectImportID reads the `uuid` from prior state and uses it as the
// import identifier; the resource has no `id` attribute.
func testAccProjectImportID(s *terraform.State) (string, error) {
	for _, m := range s.Modules {
		if r, ok := m.Resources["fco_project.test"]; ok && r.Primary != nil {
			if uuid, ok := r.Primary.Attributes["uuid"]; ok {
				return uuid, nil
			}
		}
	}
	return "", fmt.Errorf("resource fco_project.test not found in prior state")
}
