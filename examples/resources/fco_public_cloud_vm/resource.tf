resource "fco_project" "my_project" {
  name        = "my_project"
  description = "Project for my VMs"
}

resource "fco_public_cloud_vm" "my_vm" {
  project_uuid = fco_project.my_project.uuid
  name         = "my_vm"
  size         = "small"

  labels = {
    env  = "dev"
    team = "platform"
  }
}

output "vm_uuid" {
  value = fco_public_cloud_vm.my_vm.uuid
}

output "vm_status" {
  value = fco_public_cloud_vm.my_vm.status.operation
}