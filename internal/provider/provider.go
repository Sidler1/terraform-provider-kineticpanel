package provider

import (
	"context"
	"os"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ provider.Provider = &KineticpanelProvider{}

type KineticpanelProvider struct{ version string }

type kineticpanelProviderModel struct {
	Host           types.String `tfsdk:"host"`
	APIKey         types.String `tfsdk:"api_key"`
	UseApplication types.Bool   `tfsdk:"use_application"`
}

func init() {
	if strings.EqualFold(os.Getenv("KINETICPANEL_DEBUG"), "true") {
		tflog.WithLevel(hclog.Debug)
	}
}

func (p *KineticpanelProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kineticpanel"
	resp.Version = p.version
}

func (p *KineticpanelProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for Kinetic Panel (Pterodactyl-compatible). Optimized for https://kineticpanel.net",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:    true,
				Description: "Base URL of the panel. Defaults to `https://kineticpanel.net` if not set. Using other hosts may not work reliably.",
			},
			"api_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Client or Application API key.",
			},
			"use_application": schema.BoolAttribute{
				Optional:    true,
				Description: "Use Application API (admin tasks, e.g. creating servers). Default: true. Set false for Client API.",
			},
		},
	}
}

func (p *KineticpanelProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config kineticpanelProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := config.Host.ValueString()
	if host == "" {
		host = os.Getenv("KINETICPANEL_HOST")
	}
	if host == "" {
		host = "https://kineticpanel.net"
	}

	if config.Host.ValueString() != "" && !strings.EqualFold(strings.TrimRight(config.Host.ValueString(), "/"), "https://kineticpanel.net") {
		resp.Diagnostics.AddWarning("Non-standard host", "This provider is optimized for https://kineticpanel.net. Other instances may have compatibility issues.")
	}

	apiKey := config.APIKey.ValueString()
	if apiKey == "" {
		apiKey = os.Getenv("KINETICPANEL_API_KEY")
	}

	useApp := true
	if !config.UseApplication.IsNull() {
		useApp = config.UseApplication.ValueBool()
	}

	if apiKey == "" {
		resp.Diagnostics.AddError("Missing configuration", "api_key is required")
		return
	}

	client := NewClient(host, apiKey, useApp)
	tflog.Info(ctx, "Provider configured", map[string]any{"host": host, "use_application": useApp})

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *KineticpanelProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewServerResource,
		NewServerPowerResource,
		NewServerCommandResource,
		NewServerRenameResource,
		NewServerReinstallResource,
		NewServerDockerImageResource,
		NewServerStartupVariableResource,
	}
}

func (p *KineticpanelProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewServersDataSource,
		NewServerDataSource,
		NewServerUtilizationDataSource,
		NewServerStartupDataSource,
		NewServerActivityLogsDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider { return &KineticpanelProvider{version: version} }
}
