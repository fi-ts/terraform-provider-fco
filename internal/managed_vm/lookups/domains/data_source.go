package domains

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
	_ datasource.DataSource              = &DomainsDataSource{}
	_ datasource.DataSourceWithConfigure = &DomainsDataSource{}
)

// NewDomainsDataSource returns the fco_managed_vm_domains data source.
func NewDomainsDataSource() datasource.DataSource {
	return &DomainsDataSource{}
}

// DomainsDataSource lists Windows domains available to the tenant, filtered by
// datacenter location.
type DomainsDataSource struct {
	session *session.Session
}

// Metadata implements datasource.DataSource.
func (*DomainsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_vm_domains"
}

// Schema implements datasource.DataSource.
func (*DomainsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceSchema.Schema{
		MarkdownDescription: "Lists all Windows domains available to the tenant in a datacenter location.\n\n" +
			"Access the results via `fco_managed_vm_domains.<name>.domains`.",
		Attributes: domainsListDataSourceAttributes(),
	}
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *DomainsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *DomainsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data listDomainsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	locationUUID := ""
	if !data.LocationUUID.IsNull() && !data.LocationUUID.IsUnknown() {
		locationUUID = data.LocationUUID.ValueString()
	}
	r, err := d.session.Client.Apimvmv1().Domain().List(ctx, &mvmv1.DomainServiceListRequest{
		Tenant:       d.session.Tenant,
		LocationUuid: locationUUID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to read domains", err.Error())
		return
	}
	domains := r.GetDomains()
	data.Domains = make([]domainView, 0, len(domains))
	ids := make([]string, 0, len(domains))
	for _, domain := range domains {
		data.Domains = append(data.Domains, domainViewFromAPI(domain))
		ids = append(ids, domain.GetUuid())
	}
	data.ID = shared.ListID(ids)
	tflog.Trace(ctx, "read domains", map[string]any{"count": len(domains)})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
