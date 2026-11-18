# List the projects of the tenant the provider is configured with.
data "fco_projects" "all" {}

# Optionally filter by name and/or labels.
data "fco_projects" "dev" {
  labels = {
    env = "dev"
  }
}

output "projects" {
  value = data.fco_projects.all.projects
}
