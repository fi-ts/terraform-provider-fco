package pcvm

import (
	"context"

	pcvmv1 "github.com/fi-ts/api/go/fits/api/pcvm/v1"
	"github.com/fi-ts/terraform-provider-fco/internal/session"
	"github.com/fi-ts/terraform-provider-fco/internal/shared"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &PublicCloudVMDataSource{}
	_ datasource.DataSourceWithConfigure = &PublicCloudVMDataSource{}
)

// NewPublicCloudVMDataSource returns the singular fco_public_cloud_vm data source.
func NewPublicCloudVMDataSource() datasource.DataSource {
	return &PublicCloudVMDataSource{}
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

	vm, err := d.getVM(ctx, data.ProjectUUID.ValueString(), data.UUID.ValueString())
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
	data.Status = model.Status
	data.UpdatedAt = model.UpdatedAt

	tflog.Trace(ctx, "read public cloud VM", map[string]any{"uuid": data.UUID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// getVM fetches a single public cloud VM by its UUID.
func (d *PublicCloudVMDataSource) getVM(ctx context.Context, projectUUID, uuid string) (*pcvmv1.PublicCloudVM, error) {
	r, err := d.session.Client.Apipcvmv1().PublicCloudVM().Get(ctx, &pcvmv1.PublicCloudVMServiceGetRequest{
		Uuid:        uuid,
		ProjectUuid: projectUUID,
	})
	if err != nil {
		return nil, err
	}
	return r.Vm, nil
}
