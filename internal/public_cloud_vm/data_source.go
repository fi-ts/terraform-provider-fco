package publiccloudvm

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
	_ datasource.DataSource              = &PublicCloudVMDataSource{}
	_ datasource.DataSourceWithConfigure = &PublicCloudVMDataSource{}
	_ datasource.DataSource              = &PublicCloudVMsDataSource{}
	_ datasource.DataSourceWithConfigure = &PublicCloudVMsDataSource{}
)

// NewPublicCloudVMDataSource returns the singular fco_public_cloud_vm data source.
func NewPublicCloudVMDataSource() datasource.DataSource {
	return &PublicCloudVMDataSource{}
}

// NewPublicCloudVMsDataSource returns the list fco_public_cloud_vms data source.
func NewPublicCloudVMsDataSource() datasource.DataSource {
	return &PublicCloudVMsDataSource{}
}

// PublicCloudVMDataSource looks up a single public cloud VM by its UUID.
type PublicCloudVMDataSource struct {
	session *session.Session
}

// Metadata implements datasource.DataSource.
func (*PublicCloudVMDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_public_cloud_vm"
}

// Schema implements datasource.DataSource.
func (*PublicCloudVMDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up a public cloud VM by its UUID.",
		Attributes:          publicCloudVMDataSourceAttributes(),
	}
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *PublicCloudVMDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	s, diag := shared.SessionFrom(req.ProviderData)
	resp.Diagnostics.Append(diag)
	if diag != nil {
		return
	}
	d.session = s
}

// Read implements datasource.DataSource.
func (d *PublicCloudVMDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data publicCloudVMModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vm, err := d.getVM(ctx, data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read public cloud VM", err.Error())
		return
	}

	model := publicCloudVMFromAPI(vm)
	data.UUID = model.UUID
	data.Name = model.Name
	data.Size = model.Size
	data.Labels = model.Labels
	data.Tenant = model.Tenant
	data.ProjectUUID = model.ProjectUUID
	data.StatusState = model.StatusState
	data.StatusOperation = model.StatusOperation
	data.StatusInfo = model.StatusInfo
	data.UpdatedAt = model.UpdatedAt

	tflog.Trace(ctx, "read public cloud VM", map[string]any{"uuid": data.UUID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// getVM fetches a single public cloud VM by its UUID.
func (d *PublicCloudVMDataSource) getVM(ctx context.Context, uuid string) (*pcvmv1.PublicCloudVM, error) {
	r, err := d.session.Client.Apipcvmv1().PublicCloudVM().Get(ctx, &pcvmv1.PublicCloudVMServiceGetRequest{
		Uuid:        uuid,
		ProjectUuid: d.session.ProjectUUID,
	})
	if err != nil {
		return nil, err
	}
	return r.Vm, nil
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
		ProjectUuid: d.session.ProjectUUID,
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
