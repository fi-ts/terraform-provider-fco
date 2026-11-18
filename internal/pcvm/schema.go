// Package pcvm implements the fco_public_cloud_vm resource and the
// fco_public_cloud_vm / fco_public_cloud_vms data sources on top of the
// fits.api.pcvm.v1.PublicCloudVMService.
package pcvm

import (
	"github.com/fi-ts/terraform-provider-fco/internal/shared"
	dataschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// vmAttributeDescriptions holds the shared MarkdownDescription for
// attributes described identically across the resource and data sources;
// surface-specific attributes are described inline at each usage site.
var vmAttributeDescriptions = map[string]string{
	"tenant":       "The tenant the public cloud VM belongs to.",
	"project_uuid": "The UUID of the project the public cloud VM belongs to.",
	"status":       "The status of the public cloud VM.",
}

// statusDescriptions holds the descriptions of the nested status attributes.
var statusDescriptions = map[string]string{
	"state":     "The health state of the public cloud VM: `healthy`, `warning` or `error`.",
	"operation": "The current operation of the public cloud VM: `creating`, `active`, `updating`, `deleted` or `error`.",
	"info":      "Free-text details about the current status of the public cloud VM.",
}

func publicCloudVMResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The name of the public cloud VM.",
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
			MarkdownDescription: shared.LabelsDescription,
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
			Required:            true,
			MarkdownDescription: vmAttributeDescriptions["project_uuid"] + " Changing it destroys and recreates the public cloud VM.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"status": schema.SingleNestedAttribute{
			Computed:            true,
			MarkdownDescription: vmAttributeDescriptions["status"],
			Attributes: map[string]schema.Attribute{
				"state":     schema.StringAttribute{Computed: true, MarkdownDescription: statusDescriptions["state"]},
				"operation": schema.StringAttribute{Computed: true, MarkdownDescription: statusDescriptions["operation"]},
				"info":      schema.StringAttribute{Computed: true, MarkdownDescription: statusDescriptions["info"]},
			},
		},
		"updated_at": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: shared.UpdatedAtDescription,
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
			MarkdownDescription: shared.LabelsDescription,
		},
		"status": dataschema.SingleNestedAttribute{
			Computed:            true,
			MarkdownDescription: vmAttributeDescriptions["status"],
			Attributes: map[string]dataschema.Attribute{
				"state":     dataschema.StringAttribute{Computed: true, MarkdownDescription: statusDescriptions["state"]},
				"operation": dataschema.StringAttribute{Computed: true, MarkdownDescription: statusDescriptions["operation"]},
				"info":      dataschema.StringAttribute{Computed: true, MarkdownDescription: statusDescriptions["info"]},
			},
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
		Required:            true,
		MarkdownDescription: "The UUID of the project the public cloud VM belongs to.",
	}
	attrs["updated_at"] = dataschema.StringAttribute{
		Computed:            true,
		MarkdownDescription: shared.UpdatedAtDescription,
	}

	return attrs
}

func publicCloudVMListDataSourceAttributes() map[string]dataschema.Attribute {
	return map[string]dataschema.Attribute{
		"id": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "A content hash identifying the current set of listed public cloud VMs.",
		},
		"project_uuid": dataschema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The UUID of the project to list public cloud VMs from.",
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
