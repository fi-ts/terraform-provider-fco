# List all public cloud VMs in a project.
data "fco_public_cloud_vms" "all" {
  project_uuid = fco_project.my_project.uuid
}

# Filter by name.
data "fco_public_cloud_vms" "by_name" {
  project_uuid = fco_project.my_project.uuid
  name         = "my_vm"
}

# Filter by labels.
data "fco_public_cloud_vms" "by_label" {
  project_uuid = fco_project.my_project.uuid
  labels = {
    env = "dev"
  }
}

# access with `fco_public_cloud_vms.all.vms`