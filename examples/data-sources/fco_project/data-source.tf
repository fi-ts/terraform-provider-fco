# Look up a single project by UUID.
data "fco_project" "example" {
  uuid = "00000000-0000-0000-0000-000000000001"
}

output "project" {
  value = data.fco_project.example
}
