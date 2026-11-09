package domains

import (
	dataschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func domainsListDataSourceAttributes() map[string]dataschema.Attribute {
	return map[string]dataschema.Attribute{
		"id": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "A content hash identifying the current set of listed domains.",
		},
		"location_uuid": dataschema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The UUID of the datacenter location to list domains for. This is required by the API.",
		},
		"domains": dataschema.ListNestedAttribute{
			Computed:            true,
			MarkdownDescription: "The list of Windows domains available to the tenant in the given location.",
			NestedObject: dataschema.NestedAttributeObject{
				Attributes: map[string]dataschema.Attribute{
					"uuid": dataschema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The UUID of the Windows domain.",
					},
					"fqdn": dataschema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The FQDN of the Windows domain.",
					},
					"tenant": dataschema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The tenant this Windows domain belongs to.",
					},
				},
			},
		},
	}
}
