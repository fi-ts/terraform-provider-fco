# List all operating systems available to the tenant.
data "fco_managed_vm_operating_systems" "all" {
}

# Access with `fco_managed_vm_operating_systems.all.operating_systems`.
# e.g. fco_managed_vm_operating_systems.all.operating_systems[0].uuid
