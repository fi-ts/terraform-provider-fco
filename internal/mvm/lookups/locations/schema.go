package locations

import (
	dataschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func locationsListDataSourceAttributes() map[string]dataschema.Attribute {
	return map[string]dataschema.Attribute{
		"id": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "A content hash identifying the current set of listed locations.",
		},
		"locations": dataschema.ListNestedAttribute{
			Computed:            true,
			MarkdownDescription: "The list of all available datacenter locations.",
			NestedObject: dataschema.NestedAttributeObject{
				Attributes: map[string]dataschema.Attribute{
					"uuid": dataschema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The UUID of the datacenter location.",
					},
					"title": dataschema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The title of the datacenter location, e.g. `Nürnberg (RZ1)`.",
					},
				},
			},
		},
	}
}
