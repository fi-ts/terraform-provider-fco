package operatingsystems

import (
	"context"

	mvmv1 "github.com/fi-ts/api/go/fits/api/mvm/v1"
	"github.com/fi-ts/terraform-provider-fco/internal/session"
	"github.com/fi-ts/terraform-provider-fco/internal/shared"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &OperatingSystemsDataSource{}
	_ datasource.DataSourceWithConfigure = &OperatingSystemsDataSource{}
)

// NewOperatingSystemsDataSource returns the fco_managed_vm_operating_systems data source.
func NewOperatingSystemsDataSource() datasource.DataSource {
	return &OperatingSystemsDataSource{}
}

// OperatingSystemsDataSource lists operating systems available to the tenant.
type OperatingSystemsDataSource struct {
	session *session.Session
}

// Metadata implements datasource.DataSource.
func (*OperatingSystemsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_vm_operating_systems"
}

// Schema implements datasource.DataSource.
func (*OperatingSystemsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceSchema.Schema{
		MarkdownDescription: "Lists all operating systems available to the tenant.\n\n" +
			"Access the results via `fco_managed_vm_operating_systems.<name>.operating_systems`.",
		Attributes: operatingSystemsListDataSourceAttributes(),
	}
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *OperatingSystemsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Configure is also invoked during the validate walk, before the provider
	// has been configured; there is no provider data yet, so do nothing.
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
func (d *OperatingSystemsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data listOperatingSystemsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r, err := d.session.Client.Apimvmv1().OS().List(ctx, &mvmv1.OSServiceListRequest{
		Tenant: d.session.Tenant,
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to read operating systems", err.Error())
		return
	}
	oses := r.GetOperatingSystems()
	data.OperatingSystems = make([]operatingSystemView, 0, len(oses))
	ids := make([]string, 0, len(oses))
	for _, os := range oses {
		data.OperatingSystems = append(data.OperatingSystems, operatingSystemViewFromAPI(os))
		ids = append(ids, os.GetUuid())
	}
	data.ID = shared.ListID(ids)
	tflog.Trace(ctx, "read operating systems", map[string]any{"count": len(oses)})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
