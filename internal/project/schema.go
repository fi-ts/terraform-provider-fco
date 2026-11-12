package project

import (
	dataschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// projectAttributeDescriptions holds the shared MarkdownDescription for
// attributes described identically across the resource and data sources;
// surface-specific attributes are described inline at each usage site.
var projectAttributeDescriptions = map[string]string{
	"name":        "The name of the project. Must be between 2 and 128 characters and unique per tenant. The name is immutable; changing it destroys and recreates the project.",
	"description": "A free-text description of the project. Must be shorter than 256 characters.",
	"avatar_url":  "The URL of the project avatar image.",
	"labels":      "A set of key/value label pairs assigned to the project.",
	"tenant":      "The tenant the project belongs to.",
	"created_at":  "The timestamp (RFC 3339) when the project was created.",
}

func projectResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: projectAttributeDescriptions["name"],
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"description": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: projectAttributeDescriptions["description"],
		},
		"avatar_url": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: projectAttributeDescriptions["avatar_url"],
		},
		"labels": schema.MapAttribute{
			Optional:            true,
			ElementType:         types.StringType,
			MarkdownDescription: projectAttributeDescriptions["labels"],
		},
		"uuid": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The unique identifier (UUID) of the project.",
		},
		"tenant": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: projectAttributeDescriptions["tenant"],
		},
		"created_at": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: projectAttributeDescriptions["created_at"],
		},
		"updated_at": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Server-side last-modified timestamp (RFC 3339). Managed automatically for optimistic locking.",
		},
	}
}

// projectViewAttributes is the read-only attribute set mirroring projectView,
// shared by the fco_projects list items and the fco_project data source. It
// cannot be reused for the resource, which uses a different
// required/computed/optional split.
func projectViewAttributes() map[string]dataschema.Attribute {
	return map[string]dataschema.Attribute{
		"uuid": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The UUID of the project.",
		},
		"name": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: projectAttributeDescriptions["name"],
		},
		"description": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: projectAttributeDescriptions["description"],
		},
		"avatar_url": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: projectAttributeDescriptions["avatar_url"],
		},
		"labels": dataschema.MapAttribute{
			Computed:            true,
			ElementType:         types.StringType,
			MarkdownDescription: projectAttributeDescriptions["labels"],
		},
		"tenant": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: projectAttributeDescriptions["tenant"],
		},
		"created_at": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: projectAttributeDescriptions["created_at"],
		},
		"updated_at": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Server-side last-modified timestamp (RFC 3339).",
		},
	}
}

// projectDataSourceAttributes is the singular fco_project data source schema.
// It reuses the read-only view attributes and marks `uuid` as the required
// lookup key.
func projectDataSourceAttributes() map[string]dataschema.Attribute {
	attrs := projectViewAttributes()

	attrs["uuid"] = dataschema.StringAttribute{
		Required:            true,
		MarkdownDescription: "The UUID of the project to look up.",
	}

	return attrs
}

func projectsListDataSourceAttributes() map[string]dataschema.Attribute {
	return map[string]dataschema.Attribute{
		"id": dataschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "A content hash identifying the current set of listed projects.",
		},
		"name": dataschema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Only return projects whose name matches this value.",
		},
		"labels": dataschema.MapAttribute{
			Optional:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "Only return projects carrying all of these labels.",
		},
		"projects": dataschema.ListNestedAttribute{
			Computed:            true,
			MarkdownDescription: "The list of projects matching the given filters, scoped to the provider's tenant.",
			NestedObject: dataschema.NestedAttributeObject{
				Attributes: projectViewAttributes(),
			},
		},
	}
}
