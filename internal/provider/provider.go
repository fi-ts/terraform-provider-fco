package provider

import (
	"context"
	"log/slog"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	client "github.com/fi-ts/api/go/client"
	"github.com/fi-ts/terraform-provider-fco/internal/public_cloud_vm"
	"github.com/fi-ts/terraform-provider-fco/internal/session"
	"github.com/fi-ts/terraform-provider-fco/internal/shared"
)

// Ensure FCOProvider satisfies the provider.Provider interface.
var _ provider.Provider = &FCOProvider{}

type FCOProvider struct {
	version string
}

type FCOProviderModel struct {
	Token       types.String `tfsdk:"token"`
	BaseURL     types.String `tfsdk:"base_url"`
	ProjectUUID types.String `tfsdk:"project_uuid"`
	Tenant      types.String `tfsdk:"tenant"`
}

func (p *FCOProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "fco"
	resp.Version = p.version
}

func (p *FCOProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `FCO` provider manages resources on the fi-ts.io platform.\n\n" +
			"All provider defaults can be derived from the `FCO_*` environment variables or set in the Terraform provider configuration. " +
			"Values set in the configuration take precedence over environment variables.",
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				MarkdownDescription: "The API authentication token. Defaults to the `FCO_TOKEN` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "The base URL of the fits-api server. Defaults to the `FCO_BASE_URL` environment variable.",
				Optional:            true,
			},
			"project_uuid": schema.StringAttribute{
				MarkdownDescription: "The project UUID to operate in. Defaults to the `FCO_PROJECT_UUID` environment variable.",
				Optional:            true,
			},
			"tenant": schema.StringAttribute{
				MarkdownDescription: "The tenant login. Defaults to the `FCO_TENANT` environment variable.",
				Optional:            true,
			},
		},
	}
}

func (p *FCOProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data FCOProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values win over the FCO_* environment fallbacks.
	token := os.Getenv("FCO_TOKEN")
	if !data.Token.IsNull() {
		token = data.Token.ValueString()
	}

	baseURL := os.Getenv("FCO_BASE_URL")
	if !data.BaseURL.IsNull() {
		baseURL = data.BaseURL.ValueString()
	}

	projectUUID := os.Getenv("FCO_PROJECT_UUID")
	if !data.ProjectUUID.IsNull() {
		projectUUID = data.ProjectUUID.ValueString()
	}

	tenant := os.Getenv("FCO_TENANT")
	if !data.Tenant.IsNull() {
		tenant = data.Tenant.ValueString()
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing FCO API token",
			"The provider cannot create the FCO API client as there is an empty or missing configuration value for the FCO API token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the FCO_TOKEN environment variable.",
		)
	}

	if baseURL == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Missing FCO base URL",
			"The provider cannot create the FCO API client as there is an empty or missing configuration value for the FCO base URL. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the FCO_BASE_URL environment variable.",
		)
	}

	if projectUUID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("project_uuid"),
			"Missing FCO project UUID",
			"The provider cannot create the FCO API client as there is an empty or missing configuration value for the FCO project UUID. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the FCO_PROJECT_UUID environment variable.",
		)
	}

	if tenant == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("tenant"),
			"Missing FCO tenant",
			"The provider cannot create the FCO API client as there is an empty or missing configuration value for the FCO tenant. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the FCO_TENANT environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	debugLevel := slog.LevelInfo
	if shared.Debug {
		debugLevel = slog.LevelDebug
	}

	apiClient, err := client.New(&client.DialConfig{
		BaseURL:   baseURL,
		Token:     token,
		UserAgent: "terraform-provider-fco/" + p.version,
		Log:       slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: debugLevel})),
	})
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Invalid FCO API token",
			err.Error(),
		)

		return
	}

	sess := &session.Session{
		Client:      apiClient,
		ProjectUUID: projectUUID,
		Tenant:      tenant,
	}

	resp.DataSourceData = sess
	resp.ResourceData = sess
}

func (p *FCOProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		publiccloudvm.NewPublicCloudVMResource,
	}
}

func (p *FCOProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		publiccloudvm.NewPublicCloudVMDataSource,
		publiccloudvm.NewPublicCloudVMsDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &FCOProvider{
			version: version,
		}
	}
}
