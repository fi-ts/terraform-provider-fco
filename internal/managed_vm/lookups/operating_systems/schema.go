package operatingsystems

import (
	dataschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func operatingSystemsListDataSourceAttributes() map[string]dataschema.Attribute {
	return map[string]dataschema.Attribute{
		"id": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "A content hash identifying the current set of listed operating systems.",
		},
		"operating_systems": dataschema.ListNestedAttribute{
			Computed:            true,
			MarkdownDescription: "The list of all operating systems available to the tenant.",
			NestedObject: dataschema.NestedAttributeObject{
				Attributes: map[string]dataschema.Attribute{
					"uuid": dataschema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The UUID of the operating system.",
					},
					"title": dataschema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The title of the operating system.",
					},
					"type": dataschema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The type of the operating system: `linux` or `windows`.",
					},
				},
			},
		},
	}
}
