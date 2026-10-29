// Package publiccloudvm implements the fco_public_cloud_vm resource and the
// fco_public_cloud_vm / fco_public_cloud_vms data sources on top of the
// fits.api.pcvm.v1.PublicCloudVMService.
package publiccloudvm

import (
	dataschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// vmAttributeDescriptions holds the shared MarkdownDescription for
// attributes described identically across the resource and data sources;
// surface-specific attributes are described inline at each usage site.
var vmAttributeDescriptions = map[string]string{
	"labels":           "A set of key/value label pairs assigned to the public cloud VM.",
	"tenant":           "The tenant the public cloud VM belongs to.",
	"project_uuid":     "The UUID of the project the public cloud VM belongs to.",
	"status_state":     "The health state of the public cloud VM: `healthy`, `warning` or `error`.",
	"status_operation": "The current operation of the public cloud VM: `creating`, `active`, `updating`, `deleted` or `error`.",
	"status_info":      "Free-text details about the current status of the public cloud VM.",
}

func publicCloudVMResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The name of the public cloud VM. Must be between 2 and 128 characters.",
		},
		"size": schema.StringAttribute{
			Required: true,
			MarkdownDescription: "The size of the public cloud VM. Must be one of `small`, `medium` or `large`. " +
				"Only upgrades (new size >= current size) are allowed; downgrades are rejected.",
			Validators: sizeValidator,
		},
		"labels": schema.MapAttribute{
			Optional:            true,
			ElementType:         types.StringType,
			MarkdownDescription: vmAttributeDescriptions["labels"],
		},
		"uuid": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The unique identifier (UUID) of the public cloud VM.",
		},
		"tenant": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: vmAttributeDescriptions["tenant"],
		},
		"project_uuid": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: vmAttributeDescriptions["project_uuid"],
		},
		"status_state": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: vmAttributeDescriptions["status_state"],
		},
		"status_operation": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: vmAttributeDescriptions["status_operation"],
		},
		"status_info": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: vmAttributeDescriptions["status_info"],
		},
		"updated_at": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Server-side last-modified timestamp (RFC 3339). Managed automatically for optimistic locking.",
		},
	}
}

// publicCloudVMViewAttributes is the read-only attribute set mirroring
// publicCloudVMView, shared by the fco_public_cloud_vms list items and the
// singular data source. It cannot be reused for the resource, which uses a
// different required/computed split.
func publicCloudVMViewAttributes() map[string]dataschema.Attribute {
	return map[string]dataschema.Attribute{
		"uuid": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The UUID of the public cloud VM.",
		},
		"name": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The name of the public cloud VM.",
		},
		"size": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The size of the public cloud VM.",
		},
		"labels": dataschema.MapAttribute{
			Computed:            true,
			ElementType:         types.StringType,
			MarkdownDescription: vmAttributeDescriptions["labels"],
		},
		"status_state": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: vmAttributeDescriptions["status_state"],
		},
		"status_operation": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: vmAttributeDescriptions["status_operation"],
		},
		"status_info": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: vmAttributeDescriptions["status_info"],
		},
	}
}

// publicCloudVMDataSourceAttributes is the singular fco_public_cloud_vm data
// source schema. It reuses the read-only view attributes (where name is already
// Computed) and marks uuid as the required lookup key.
func publicCloudVMDataSourceAttributes() map[string]dataschema.Attribute {
	attrs := publicCloudVMViewAttributes()

	attrs["uuid"] = dataschema.StringAttribute{
		Required:            true,
		MarkdownDescription: "The UUID of the public cloud VM to look up.",
	}
	attrs["tenant"] = dataschema.StringAttribute{
		Computed:            true,
		MarkdownDescription: vmAttributeDescriptions["tenant"],
	}
	attrs["project_uuid"] = dataschema.StringAttribute{
		Computed:            true,
		MarkdownDescription: vmAttributeDescriptions["project_uuid"],
	}
	attrs["updated_at"] = dataschema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "Server-side last-modified timestamp (RFC 3339).",
	}

	return attrs
}

func publicCloudVMListDataSourceAttributes() map[string]dataschema.Attribute {
	return map[string]dataschema.Attribute{
		"id": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "A content hash identifying the current set of listed public cloud VMs.",
		},
		"name": dataschema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Only return public cloud VMs whose name matches this value.",
		},
		"labels": dataschema.MapAttribute{
			Optional:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "Only return public cloud VMs carrying all of these labels.",
		},
		"vms": dataschema.ListNestedAttribute{
			Computed:            true,
			MarkdownDescription: "The list of public cloud VMs matching the given filters.",
			NestedObject: dataschema.NestedAttributeObject{
				Attributes: publicCloudVMViewAttributes(),
			},
		},
	}
}
