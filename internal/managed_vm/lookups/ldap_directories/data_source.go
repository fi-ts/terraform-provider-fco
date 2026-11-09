package ldapdirectories

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
	_ datasource.DataSource              = &LDAPDirectoriesDataSource{}
	_ datasource.DataSourceWithConfigure = &LDAPDirectoriesDataSource{}
)

// NewLDAPDirectoriesDataSource returns the fco_managed_vm_ldap_directories data source.
func NewLDAPDirectoriesDataSource() datasource.DataSource {
	return &LDAPDirectoriesDataSource{}
}

// LDAPDirectoriesDataSource lists LDAP directories available to the tenant.
type LDAPDirectoriesDataSource struct {
	session *session.Session
}

// Metadata implements datasource.DataSource.
func (*LDAPDirectoriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_vm_ldap_directories"
}

// Schema implements datasource.DataSource.
func (*LDAPDirectoriesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceSchema.Schema{
		MarkdownDescription: "Lists all LDAP directories available to the tenant.\n\n" +
			"Access the results via `fco_managed_vm_ldap_directories.<name>.ldap_directories`.",
		Attributes: ldapDirectoriesListDataSourceAttributes(),
	}
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *LDAPDirectoriesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *LDAPDirectoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data listLDAPDirectoriesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r, err := d.session.Client.Apimvmv1().LDAP().List(ctx, &mvmv1.LDAPServiceListRequest{
		Tenant: d.session.Tenant,
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to read LDAP directories", err.Error())
		return
	}
	ldaps := r.GetLdaps()
	data.LDAPDirectories = make([]ldapDirectoryView, 0, len(ldaps))
	ids := make([]string, 0, len(ldaps))
	for _, ldap := range ldaps {
		data.LDAPDirectories = append(data.LDAPDirectories, ldapDirectoryViewFromAPI(ldap))
		ids = append(ids, ldap.GetUuid())
	}
	data.ID = shared.ListID(ids)
	tflog.Trace(ctx, "read LDAP directories", map[string]any{"count": len(ldaps)})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
