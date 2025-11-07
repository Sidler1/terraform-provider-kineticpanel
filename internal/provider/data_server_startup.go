package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ServerStartupDataSource{}

// ServerStartupDataSource fetches startup command and variables for a server.
type ServerStartupDataSource struct {
	client *Client
}

// startupModel holds the data source state.
type startupModel struct {
	ServerID       types.String `tfsdk:"server_id"`
	StartupCommand types.String `tfsdk:"startup_command"`
	EggID          types.Int64  `tfsdk:"egg_id"`
	DockerImage    types.String `tfsdk:"docker_image"`
	// Dynamic list of environment variables
	Environment types.Map `tfsdk:"environment"`
}

func NewServerStartupDataSource() datasource.DataSource {
	return &ServerStartupDataSource{}
}

func (d *ServerStartupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_startup"
}

func (d *ServerStartupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches startup command and environment variables for a Kinetic Panel server (Client API).",
		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				Required:    true,
				Description: "Short server identifier (e.g. `abc123`).",
			},
			"startup_command": schema.StringAttribute{
				Computed:    true,
				Description: "Full startup command used by the server.",
			},
			"egg_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Egg ID this server is based on.",
			},
			"docker_image": schema.StringAttribute{
				Computed:    true,
				Description: "Docker image used for the container.",
			},
			"environment": schema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Map of environment variables passed to the startup process.",
			},
		},
	}
}

func (d *ServerStartupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *ServerStartupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config struct {
		ServerID types.String `tfsdk:"server_id"`
	}
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverID := config.ServerID.ValueString()
	path := "/servers/" + serverID + "/startup"

	body, err := d.client.Get(path)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to fetch startup for server %s: %v", serverID, err))
		return
	}

	var apiResp struct {
		StartupCommand string            `json:"startup"`
		Egg            int64             `json:"egg"`
		DockerImage    string            `json:"image"`
		Environment    map[string]string `json:"environment"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", err.Error())
		return
	}

	// Convert map[string]string → types.Map
	envMap := make(map[string]types.String)
	for k, v := range apiResp.Environment {
		envMap[k] = types.StringValue(v)
	}
	env, diags := types.MapValueFrom(ctx, types.StringType, envMap)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := startupModel{
		ServerID:       config.ServerID,
		StartupCommand: types.StringValue(apiResp.StartupCommand),
		EggID:          types.Int64Value(apiResp.Egg),
		DockerImage:    types.StringValue(apiResp.DockerImage),
		Environment:    env,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
