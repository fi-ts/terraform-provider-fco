# List Windows domains available to the tenant in a datacenter location.
# `location_uuid` is required.
data "fco_managed_vm_domains" "in_location" {
  location_uuid = data.fco_managed_vm_locations.all.locations[0].uuid
}

# Access with `fco_managed_vm_domains.in_location.domains`.
# e.g. fco_managed_vm_domains.in_location.domains[0].uuid
