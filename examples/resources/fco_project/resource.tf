# Manage a project within the tenant the provider is configured with.
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