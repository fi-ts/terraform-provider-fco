# Look up a public cloud VM by its UUID.
data "fco_public_cloud_vm" "by_uuid" {
  uuid = fco_public_cloud_vm.my_vm.uuid
}