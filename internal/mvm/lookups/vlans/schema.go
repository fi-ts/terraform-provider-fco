package vlans

import (
	dataschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func vlansListDataSourceAttributes() map[string]dataschema.Attribute {
	return map[string]dataschema.Attribute{
		"id": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "A content hash identifying the current set of listed VLANs.",
		},
		"location_uuid": dataschema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The UUID of the datacenter location to list VLANs for. This is required by the API.",
		},
		"domain_uuid": dataschema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Only return VLANs associated with this Windows domain.",
		},
		"vlans": dataschema.ListNestedAttribute{
			Computed:            true,
			MarkdownDescription: "The list of VLANs available to the tenant in the given location, optionally filtered by `domain_uuid`.",
			NestedObject: dataschema.NestedAttributeObject{
				Attributes: map[string]dataschema.Attribute{
					"uuid": dataschema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The UUID of the VLAN.",
					},
					"vlan_id": dataschema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "The VLAN ID a network interface is attached to.",
					},
					"subnet_title": dataschema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The title of the subnet this VLAN belongs to, including the stage type.",
					},
					"subnet_cidr": dataschema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The CIDR of the subnet this VLAN belongs to.",
					},
					"location_uuid": dataschema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The UUID of the location this VLAN belongs to.",
					},
					"tenant": dataschema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The tenant this VLAN belongs to.",
					},
					"domain_uuid": dataschema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The UUID of the Windows domain this VLAN is associated with, if any.",
					},
				},
			},
		},
	}
}
