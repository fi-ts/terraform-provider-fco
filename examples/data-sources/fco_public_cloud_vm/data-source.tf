# Look up a public cloud VM by its UUID.
data "fco_public_cloud_vm" "by_uuid" {
  project_uuid = fco_public_cloud_vm.my_vm.project_uuid
  uuid         = fco_public_cloud_vm.my_vm.uuid
}