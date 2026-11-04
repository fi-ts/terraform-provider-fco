# List VLANs available to the tenant in a datacenter location.
# `location_uuid` is required; `domain_uuid` is an optional filter.
data "fco_managed_vm_vlans" "in_location" {
  location_uuid = data.fco_managed_vm_locations.all.locations[0].uuid
}

# Access with `fco_managed_vm_vlans.in_location.vlans`.
# e.g. fco_managed_vm_vlans.in_location.vlans[0].uuid
