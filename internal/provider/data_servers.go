package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ServerDataSource{}

// ServerDataSource reads a single server using the Client API.
type ServerDataSource struct {
	client *Client
}

// serverDataModel represents the data source schema.
type serverDataModel struct {
	ServerID       types.String `tfsdk:"server_id"`
	ID             types.String `tfsdk:"id"`
	Identifier     types.String `tfsdk:"identifier"`
	InternalID     types.Int64  `tfsdk:"internal_id"`
	Name           types.String `tfsdk:"name"`
	UUID           types.String `tfsdk:"uuid"`
	Description    types.String `tfsdk:"description"`
	Suspended      types.Bool   `tfsdk:"suspended"`
	Owner          types.Bool   `tfsdk:"owner"`
	Node           types.String `tfsdk:"node"`
	DockerImage    types.String `tfsdk:"docker_image"`
	StartupCommand types.String `tfsdk:"invocation"`
	Memory         types.Int64  `tfsdk:"memory"`
	Disk           types.Int64  `tfsdk:"disk"`
	CPU            types.Int64  `tfsdk:"cpu"`
	Swap           types.Int64  `tfsdk:"swap"`
	IO             types.Int64  `tfsdk:"io"`
	AllocationID   types.Int64  `tfsdk:"allocation_id"`
	AllocationIP   types.String `tfsdk:"allocation_ip"`
	AllocationPort types.Int64  `tfsdk:"allocation_port"`
}

// NewServerDataSource creates a new data source.
func NewServerDataSource() datasource.DataSource {
	return &ServerDataSource{}
}

// Metadata sets the Terraform type name.
func (d *ServerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

// Schema defines the data source attributes.
func (d *ServerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches details of a single server from Kinetic Panel (Client API).",
		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				Required:    true,
				Description: "Short server identifier (e.g. `abc123`).",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Same as `server_id`.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Server name.",
			},
			"uuid": schema.StringAttribute{
				Computed:    true,
				Description: "Server UUID.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Server description.",
			},
			"suspended": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the server is suspended.",
			},
			"owner": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the API key belongs to the server owner.",
			},
			"node": schema.StringAttribute{
				Computed:    true,
				Description: "Node the server runs on.",
			},
			"docker_image": schema.StringAttribute{
				Computed:    true,
				Description: "Docker image used.",
			},
			"memory": schema.Int64Attribute{
				Computed:    true,
				Description: "Allocated memory in MB.",
			},
			"disk": schema.Int64Attribute{
				Computed:    true,
				Description: "Allocated disk in MB.",
			},
			"cpu": schema.Int64Attribute{
				Computed:    true,
				Description: "CPU limit in percent.",
			},
			"swap": schema.Int64Attribute{
				Computed:    true,
				Description: "Swap allocation in MB.",
			},
			"io": schema.Int64Attribute{
				Computed:    true,
				Description: "IO priority.",
			},
			"allocation_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Primary allocation ID.",
			},
			"allocation_ip": schema.StringAttribute{
				Computed:    true,
				Description: "Primary allocation IP.",
			},
			"allocation_port": schema.Int64Attribute{
				Computed:    true,
				Description: "Primary allocation port.",
			},
		},
	}
}

// Configure injects the client.
func (d *ServerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches the server and populates the state.
func (d *ServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config serverDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverID := config.ServerID.ValueString()
	path := "/servers/" + serverID

	body, err := d.client.Get(path)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to fetch server %s: %v", serverID, err))
		return
	}

	var apiResp struct {
		Attributes struct {
			Name           string `json:"name"`
			UUID           string `json:"uuid"`
			Description    string `json:"description"`
			Suspended      bool   `json:"suspended"`
			Identifier     string `json:"identifier"`
			InternalID     int64  `json:"internal_id"`
			StartupCommand string `json:"invocation"`
			Limits         struct {
				Memory int64 `json:"memory"`
				Swap   int64 `json:"swap"`
				Disk   int64 `json:"disk"`
				IO     int64 `json:"io"`
				CPU    int64 `json:"cpu"`
			} `json:"limits"`
			FeatureLimits struct {
				Databases int64 `json:"databases"`
				Backups   int64 `json:"backups"`
			} `json:"feature_limits"`
			User      int64  `json:"user"`
			Node      string `json:"node"`
			Container struct {
				Image string `json:"image"`
			} `json:"container"`
			Allocation struct {
				ID    int64  `json:"id"`
				IP    string `json:"ip"`
				Port  int64  `json:"port"`
				Alias string `json:"alias"`
			} `json:"allocation"`
			IsServerOwner bool `json:"server_owner"`
		} `json:"attributes"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", err.Error())
		return
	}

	a := apiResp.Attributes

	state := serverDataModel{
		ServerID:       config.ServerID,
		ID:             config.ServerID,
		Name:           types.StringValue(a.Name),
		UUID:           types.StringValue(a.UUID),
		Description:    types.StringValue(a.Description),
		Suspended:      types.BoolValue(a.Suspended),
		Owner:          types.BoolValue(a.IsServerOwner),
		Node:           types.StringValue(a.Node),
		DockerImage:    types.StringValue(a.Container.Image),
		Memory:         types.Int64Value(a.Limits.Memory),
		Disk:           types.Int64Value(a.Limits.Disk),
		CPU:            types.Int64Value(a.Limits.CPU),
		Swap:           types.Int64Value(a.Limits.Swap),
		IO:             types.Int64Value(a.Limits.IO),
		AllocationID:   types.Int64Value(a.Allocation.ID),
		AllocationIP:   types.StringValue(a.Allocation.IP),
		AllocationPort: types.Int64Value(a.Allocation.Port),
		Identifier:     types.StringValue(a.Identifier),
		InternalID:     types.Int64Value(a.InternalID),
		StartupCommand: types.StringValue(a.StartupCommand),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
