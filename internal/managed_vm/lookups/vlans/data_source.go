package vlans

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
	_ datasource.DataSource              = &VLANsDataSource{}
	_ datasource.DataSourceWithConfigure = &VLANsDataSource{}
)

// NewVLANsDataSource returns the fco_managed_vm_vlans data source.
func NewVLANsDataSource() datasource.DataSource {
	return &VLANsDataSource{}
}

// VLANsDataSource lists VLANs available to the tenant in a datacenter location,
// optionally filtered by Windows domain.
type VLANsDataSource struct {
	session *session.Session
}

// Metadata implements datasource.DataSource.
func (*VLANsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_vm_vlans"
}

// Schema implements datasource.DataSource.
func (*VLANsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceSchema.Schema{
		MarkdownDescription: "Lists all VLANs available to the tenant in a datacenter location, optionally filtered by `domain_uuid`.\n\n" +
			"Access the results via `fco_managed_vm_vlans.<name>.vlans`.",
		Attributes: vlansListDataSourceAttributes(),
	}
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *VLANsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *VLANsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data listVLANsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	locationUUID := ""
	if !data.LocationUUID.IsNull() && !data.LocationUUID.IsUnknown() {
		locationUUID = data.LocationUUID.ValueString()
	}
	query := &mvmv1.VLANQuery{}
	if !data.DomainUUID.IsNull() && !data.DomainUUID.IsUnknown() {
		domainUUID := data.DomainUUID.ValueString()
		query.DomainUuid = &domainUUID
	}
	r, err := d.session.Client.Apimvmv1().VLAN().List(ctx, &mvmv1.VLANServiceListRequest{
		Tenant:       d.session.Tenant,
		LocationUuid: locationUUID,
		Query:        query,
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VLANs", err.Error())
		return
	}
	vlans := r.GetVlans()
	data.VLANs = make([]vlanView, 0, len(vlans))
	ids := make([]string, 0, len(vlans))
	for _, vlan := range vlans {
		data.VLANs = append(data.VLANs, vlanViewFromAPI(vlan))
		ids = append(ids, vlan.GetUuid())
	}
	data.ID = shared.ListID(ids)
	tflog.Trace(ctx, "read VLANs", map[string]any{"count": len(vlans)})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
