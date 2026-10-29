# List all public cloud VMs in the project.
data "fco_public_cloud_vms" "all" {
}

# Filter by name.
data "fco_public_cloud_vms" "by_name" {
  name = "my_vm"
}

# Filter by labels.
data "fco_public_cloud_vms" "by_label" {
  labels = {
    env = "dev"
  }
}

# access with `fco_public_cloud_vms.all.vms`