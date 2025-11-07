package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &KineticpanelProvider{}

type KineticpanelProvider struct {
	version string
}

type kineticpanelProviderModel struct {
	Host           types.String `tfsdk:"host"`
	APIKey         types.String `tfsdk:"api_key"`
	UseApplication types.Bool   `tfsdk:"use_application"`
}

func (p *KineticpanelProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kineticpanel"
	resp.Version = p.version
}

func (p *KineticpanelProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Required:    false,
				Optional:    true,
				Description: "Base URL of your Kinetic Panel instance (include https://, no trailing slash). Example: https://kineticpanel.net",
			},
			"api_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Pterodactyl API key (Application key for creating servers, Client key for managing existing ones).",
			},
			"use_application": schema.BoolAttribute{
				Optional:    true,
				Description: "Set to false to use Client API instead of Application API. Default: true",
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

	// Allow env overrides
	host := config.Host.ValueString()
	if host == "" {
		host = "https://kineticpanel.net"
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
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *KineticpanelProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewServerResource,
	}
}

func (p *KineticpanelProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &KineticpanelProvider{version: version}
	}
}
