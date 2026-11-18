package project

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strings"

	apiv1 "github.com/fi-ts/api/go/fits/api/v1"
	"github.com/fi-ts/terraform-provider-fco/internal/session"
	"github.com/fi-ts/terraform-provider-fco/internal/shared"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &ProjectsDataSource{}
	_ datasource.DataSourceWithConfigure = &ProjectsDataSource{}
)

// NewProjectsDataSource returns the list fco_projects data source.
func NewProjectsDataSource() datasource.DataSource {
	return &ProjectsDataSource{}
}

// ProjectsDataSource lists projects of the provider's tenant, optionally
// filtered by `name` and/or `labels`.
type ProjectsDataSource struct {
	session *session.Session
}

// Metadata implements datasource.DataSource.
func (*ProjectsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects"
}

// Schema implements datasource.DataSource.
func (*ProjectsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the projects of the tenant the provider is configured with, optionally filtered by `name` and/or `labels`.\n\n" +
			"Access the results via `fco_projects.<name>.projects`.",
		Attributes: projectsListDataSourceAttributes(),
	}
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *ProjectsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProjectsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data listProjectsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listReq := &apiv1.ProjectServiceListRequest{
		Tenant: &d.session.Tenant,
	}
	if !data.Name.IsNull() && data.Name.ValueString() != "" {
		name := data.Name.ValueString()
		listReq.Name = &name
	}
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		var labels map[string]string
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(labels) > 0 {
			listReq.Labels = &apiv1.Labels{Labels: labels}
		}
	}

	r, err := d.session.Client.Apiv1().Project().List(ctx, listReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read projects", err.Error())
		return
	}
	projects := r.GetProjects()
	tflog.Trace(ctx, "read projects", map[string]any{"count": len(projects)})

	data.Projects = make([]projectView, 0, len(projects))
	ids := make([]string, 0, len(projects))
	for _, p := range projects {
		data.Projects = append(data.Projects, projectViewFromAPI(p))
		ids = append(ids, p.GetUuid())
	}
	data.ID = types.StringValue(fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(ids, ",")))))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
