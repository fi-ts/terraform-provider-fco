# List all available datacenter locations.
data "fco_managed_vm_locations" "all" {
}

# Access with `fco_managed_vm_locations.all.locations`.
# e.g. fco_managed_vm_locations.all.locations[0].uuid
