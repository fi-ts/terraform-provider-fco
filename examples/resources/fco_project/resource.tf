# Manage a project within the tenant the provider is configured with.
# No project_uuid is required — projects are tenant scoped.
resource "fco_project" "my_project" {
  name        = "my-project"
  description = "Managed by Terraform"

  labels = {
    env  = "dev"
    team = "platform"
  }
}

output "project_uuid" {
  value = fco_project.my_project.uuid
}

# Referencing the same project again via a data source.
data "fco_project" "my_project" {
  uuid = fco_project.my_project.uuid
}

# List the projects of the current tenant, filtered by name.
data "fco_projects" "dev" {
  labels = {
    env = "dev"
  }
}
