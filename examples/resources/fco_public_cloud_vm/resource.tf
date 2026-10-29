resource "fco_public_cloud_vm" "my_vm" {
  name = "my_vm"
  size = "small"

  labels = {
    env  = "dev"
    team = "platform"
  }
}

output "vm_uuid" {
  value = fco_public_cloud_vm.my_vm.uuid
}

output "vm_status" {
  value = fco_public_cloud_vm.my_vm.status_operation
}