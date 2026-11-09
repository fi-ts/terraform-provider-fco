package locations

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
	_ datasource.DataSource              = &LocationsDataSource{}
	_ datasource.DataSourceWithConfigure = &LocationsDataSource{}
)

// NewLocationsDataSource returns the fco_managed_vm_locations data source.
func NewLocationsDataSource() datasource.DataSource {
	return &LocationsDataSource{}
}

// LocationsDataSource lists the available datacenter locations.
type LocationsDataSource struct {
	session *session.Session
}

// Metadata implements datasource.DataSource.
func (*LocationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_vm_locations"
}

// Schema implements datasource.DataSource.
func (*LocationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceSchema.Schema{
		MarkdownDescription: "Lists all available datacenter locations.\n\n" +
			"Access the results via `fco_managed_vm_locations.<name>.locations`.",
		Attributes: locationsListDataSourceAttributes(),
	}
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *LocationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *LocationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data listLocationsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r, err := d.session.Client.Apimvmv1().Location().List(ctx, &mvmv1.LocationServiceListRequest{})
	if err != nil {
		resp.Diagnostics.AddError("Unable to read locations", err.Error())
		return
	}
	locations := r.GetLocations()
	data.Locations = make([]locationView, 0, len(locations))
	ids := make([]string, 0, len(locations))
	for _, location := range locations {
		data.Locations = append(data.Locations, locationViewFromAPI(location))
		ids = append(ids, location.GetUuid())
	}
	data.ID = shared.ListID(ids)
	tflog.Trace(ctx, "read locations", map[string]any{"count": len(locations)})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
