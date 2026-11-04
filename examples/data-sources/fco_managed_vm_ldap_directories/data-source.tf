# List all LDAP directories available to the tenant.
data "fco_managed_vm_ldap_directories" "all" {
}

# Access with `fco_managed_vm_ldap_directories.all.ldap_directories`.
# e.g. fco_managed_vm_ldap_directories.all.ldap_directories[0].uuid
