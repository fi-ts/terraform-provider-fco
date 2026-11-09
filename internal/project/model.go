// Package project implements the fco_project resource and the
// fco_project / fco_projects data sources on top of the
// fits.api.v1.ProjectService.
package project

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// projectView holds the attributes every representation of a project shares:
// the items of the fco_projects list data source, the fco_project data source
// lookup result, and the core of projectModel.
type projectView struct {
	UUID        types.String `tfsdk:"uuid"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	AvatarURL   types.String `tfsdk:"avatar_url"`
	Labels      types.Map    `tfsdk:"labels"`
	Tenant      types.String `tfsdk:"tenant"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// projectModel describes the Terraform state for the fco_project resource and
// the fco_project data source.
type projectModel struct {
	projectView
}

// listProjectsModel describes the Terraform state for the fco_projects data
// source.
type listProjectsModel struct {
	ID       types.String  `tfsdk:"id"`
	Name     types.String  `tfsdk:"name"`
	Labels   types.Map     `tfsdk:"labels"`
	Projects []projectView `tfsdk:"projects"`
}
