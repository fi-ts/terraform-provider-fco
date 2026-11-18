package project

import (
	apiv1 "github.com/fi-ts/api/go/fits/api/v1"
	"github.com/fi-ts/terraform-provider-fco/internal/shared"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// projectViewFromAPI maps an API project to its shared read-only view.
func projectViewFromAPI(p *apiv1.Project) projectView {
	return projectView{
		UUID:        types.StringValue(p.GetUuid()),
		Name:        types.StringValue(p.GetName()),
		Description: types.StringValue(p.GetDescription()),
		Labels:      shared.LabelsFromAPI(p.GetMeta().GetLabels()),
		Tenant:      types.StringValue(p.GetTenant()),
		CreatedAt:   types.StringValue(shared.TimestampFromAPI(p.GetMeta().GetCreatedAt())),
		UpdatedAt:   types.StringValue(shared.TimestampFromAPI(p.GetMeta().GetUpdatedAt())),
	}
}

// projectFromAPI maps an API project to the resource/data source model.
func projectFromAPI(p *apiv1.Project) projectModel {
	return projectModel{
		projectView: projectViewFromAPI(p),
	}
}
