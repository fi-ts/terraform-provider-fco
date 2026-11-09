package ldapdirectories

import (
	dataschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func ldapDirectoriesListDataSourceAttributes() map[string]dataschema.Attribute {
	return map[string]dataschema.Attribute{
		"id": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "A content hash identifying the current set of listed LDAP directories.",
		},
		"ldap_directories": dataschema.ListNestedAttribute{
			Computed:            true,
			MarkdownDescription: "The list of LDAP directories available to the tenant.",
			NestedObject: dataschema.NestedAttributeObject{
				Attributes: map[string]dataschema.Attribute{
					"uuid": dataschema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The UUID of the LDAP directory.",
					},
					"fqdn": dataschema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The FQDN of the LDAP directory.",
					},
					"tenant": dataschema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The tenant this LDAP directory belongs to.",
					},
				},
			},
		},
	}
}
