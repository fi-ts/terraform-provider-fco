package pcvm

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strings"

	pcvmv1 "github.com/fi-ts/api/go/fits/api/pcvm/v1"
	apiv1 "github.com/fi-ts/api/go/fits/api/v1"
	"github.com/fi-ts/terraform-provider-fco/internal/session"
	"github.com/fi-ts/terraform-provider-fco/internal/shared"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &PublicCloudVMsDataSource{}
	_ datasource.DataSourceWithConfigure = &PublicCloudVMsDataSource{}
)

// NewPublicCloudVMsDataSource returns the list fco_public_cloud_vms data source.
func NewPublicCloudVMsDataSource() datasource.DataSource {
	return &PublicCloudVMsDataSource{}
}

// PublicCloudVMsDataSource lists public cloud VMs in the project.
type PublicCloudVMsDataSource struct {
	session *session.Session
}

// Metadata implements datasource.DataSource.
func (*PublicCloudVMsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_public_cloud_vms"
}

// Schema implements datasource.DataSource.
func (*PublicCloudVMsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all public cloud VMs in the project, optionally filtered by `name` and/or `labels`.\n\n" +
			"Access the results via `fco_public_cloud_vms.<name>.vms`.",
		Attributes: publicCloudVMListDataSourceAttributes(),
	}
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *PublicCloudVMsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	s, diag := shared.SessionFrom(req.ProviderData)
	resp.Diagnostics.Append(diag)
	if diag != nil {
		return
	}
	d.session = s
}

// Read implements datasource.DataSource.
func (d *PublicCloudVMsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data listPublicCloudVMModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	query := &pcvmv1.PublicCloudVMQuery{}
	if !data.Name.IsNull() && data.Name.ValueString() != "" {
		name := data.Name.ValueString()
		query.Name = &name
	}
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		var labels map[string]string
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(labels) > 0 {
			query.Labels = &apiv1.Labels{Labels: labels}
		}
	}

	r, err := d.session.Client.Apipcvmv1().PublicCloudVM().List(ctx, &pcvmv1.PublicCloudVMServiceListRequest{
		ProjectUuid: data.ProjectUUID.ValueString(),
		Query:       query,
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to read public cloud VMs", err.Error())
		return
	}
	vms := r.GetVms()
	tflog.Trace(ctx, "read public cloud VMs", map[string]any{"count": len(vms)})

	data.VMs = make([]publicCloudVMView, 0, len(vms))
	ids := make([]string, 0, len(vms))
	for _, vm := range vms {
		data.VMs = append(data.VMs, publicCloudVMViewFromAPI(vm))
		ids = append(ids, vm.GetUuid())
	}
	data.ID = types.StringValue(fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(ids, ",")))))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
