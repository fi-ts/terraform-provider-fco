package project

import (
	"context"

	apiv1 "github.com/fi-ts/api/go/fits/api/v1"
	"github.com/fi-ts/terraform-provider-fco/internal/session"
	"github.com/fi-ts/terraform-provider-fco/internal/shared"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &ProjectDataSource{}
	_ datasource.DataSourceWithConfigure = &ProjectDataSource{}
)

// NewProjectDataSource returns the singular fco_project data source.
func NewProjectDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

// ProjectDataSource looks up a single project by its UUID.
type ProjectDataSource struct {
	session *session.Session
}

// Metadata implements datasource.DataSource.
func (*ProjectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema implements datasource.DataSource.
func (*ProjectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up a project by its UUID.",
		Attributes:          projectDataSourceAttributes(),
	}
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *ProjectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// ProviderData is nil during validation (before ConfigureProvider runs) and
	// is not needed until the read operation; no-op in that case.
	if req.ProviderData == nil {
		return
	}
	s, diag := shared.SessionFrom(req.ProviderData)
	resp.Diagnostics.Append(diag)
	if diag != nil {
		return
	}
	d.session = s
}

// Read implements datasource.DataSource.
func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data projectModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p, err := d.getProject(ctx, data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read project", err.Error())
		return
	}

	model := projectFromAPI(p)
	data.UUID = model.UUID
	data.Name = model.Name
	data.Description = model.Description
	data.Labels = model.Labels
	data.Tenant = model.Tenant
	data.CreatedAt = model.CreatedAt
	data.UpdatedAt = model.UpdatedAt

	tflog.Trace(ctx, "read project", map[string]any{"uuid": data.UUID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// getProject fetches a single project by its UUID.
func (d *ProjectDataSource) getProject(ctx context.Context, uuid string) (*apiv1.Project, error) {
	r, err := d.session.Client.Apiv1().Project().Get(ctx, &apiv1.ProjectServiceGetRequest{
		Project: uuid,
	})
	if err != nil {
		return nil, err
	}
	return r.GetProject(), nil
}
