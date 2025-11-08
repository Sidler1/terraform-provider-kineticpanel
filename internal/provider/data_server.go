package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ServerDataSource{}

type ServerDataSource struct {
	client *Client
}

type serverDataModel struct {
	ServerID        types.String `tfsdk:"server_id"`
	ID              types.String `tfsdk:"id"`
	Identifier      types.String `tfsdk:"identifier"`
	InternalID      types.Int64  `tfsdk:"internal_id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	Suspended       types.Bool   `tfsdk:"is_suspended"`
	Installing      types.Bool   `tfsdk:"is_installing"`
	Transferring    types.Bool   `tfsdk:"is_transferring"`
	Node            types.String `tfsdk:"node"`
	SFTPIP          types.String `tfsdk:"sftp_ip"`
	SFTPPort        types.Int64  `tfsdk:"sftp_port"`
	Invocation      types.String `tfsdk:"invocation"`
	DockerImage     types.String `tfsdk:"docker_image"`
	Memory          types.Int64  `tfsdk:"memory"`
	Disk            types.Int64  `tfsdk:"disk"`
	CPU             types.Int64  `tfsdk:"cpu"`
	Swap            types.Int64  `tfsdk:"swap"`
	IO              types.Int64  `tfsdk:"io"`
	AllocationIP    types.String `tfsdk:"allocation_ip"`
	AllocationPort  types.Int64  `tfsdk:"allocation_port"`
	Environment     types.Map    `tfsdk:"environment"`
	EggFeatures     types.List   `tfsdk:"egg_features"`
	FeatureLimits   types.Object `tfsdk:"feature_limits"`
	UserPermissions types.List   `tfsdk:"user_permissions"`
}

func NewServerDataSource() datasource.DataSource {
	return &ServerDataSource{}
}

func (d *ServerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (d *ServerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single server from Kinetic Panel (Client API).",
		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				Required:    true,
				Description: "Short server identifier (e.g. `19281aed`).",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Same as `identifier`.",
			},
			"identifier": schema.StringAttribute{
				Computed:    true,
				Description: "Short server identifier.",
			},
			"internal_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Internal numeric ID.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Server name.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Server description.",
			},
			"is_suspended": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the server is suspended.",
			},
			"is_installing": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the server is installing.",
			},
			"is_transferring": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the server is transferring.",
			},
			"node": schema.StringAttribute{
				Computed:    true,
				Description: "Node name.",
			},
			"sftp_ip": schema.StringAttribute{
				Computed:    true,
				Description: "SFTP IP.",
			},
			"sftp_port": schema.Int64Attribute{
				Computed:    true,
				Description: "SFTP port.",
			},
			"invocation": schema.StringAttribute{
				Computed:    true,
				Description: "Startup command.",
			},
			"docker_image": schema.StringAttribute{
				Computed:    true,
				Description: "Docker image.",
			},
			"memory": schema.Int64Attribute{
				Computed:    true,
				Description: "Memory limit (MB).",
			},
			"disk": schema.Int64Attribute{
				Computed:    true,
				Description: "Disk limit (MB).",
			},
			"cpu": schema.Int64Attribute{
				Computed:    true,
				Description: "CPU limit (%).",
			},
			"swap": schema.Int64Attribute{
				Computed:    true,
				Description: "Swap limit (MB).",
			},
			"io": schema.Int64Attribute{
				Computed:    true,
				Description: "IO limit.",
			},
			"allocation_ip": schema.StringAttribute{
				Computed:    true,
				Description: "Primary allocation IP.",
			},
			"allocation_port": schema.Int64Attribute{
				Computed:    true,
				Description: "Primary allocation port.",
			},
			"environment": schema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Startup environment variables.",
			},
			"egg_features": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Enabled egg features.",
			},
			"feature_limits": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"databases":   schema.Int64Attribute{Computed: true},
					"allocations": schema.Int64Attribute{Computed: true},
					"backups":     schema.Int64Attribute{Computed: true},
				},
			},
			"user_permissions": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "User permissions.",
			},
		},
	}
}

func (d *ServerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *Client, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *ServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config struct {
		ServerID types.String `tfsdk:"server_id"`
	}
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
			Identifier     string `json:"identifier"`
			InternalID     int64  `json:"internal_id"`
			Name           string `json:"name"`
			Description    string `json:"description"`
			IsSuspended    bool   `json:"is_suspended"`
			IsInstalling   bool   `json:"is_installing"`
			IsTransferring bool   `json:"is_transferring"`
			Node           string `json:"node"`
			SFTPDetails    struct {
				IP   string `json:"ip"`
				Port int64  `json:"port"`
			} `json:"sftp_details"`
			Invocation    string   `json:"invocation"`
			DockerImage   string   `json:"docker_image"`
			EggFeatures   []string `json:"egg_features"`
			FeatureLimits struct {
				Databases   int64 `json:"databases"`
				Allocations int64 `json:"allocations"`
				Backups     int64 `json:"backups"`
			} `json:"feature_limits"`
			Limits struct {
				Memory int64 `json:"memory"`
				Swap   int64 `json:"swap"`
				Disk   int64 `json:"disk"`
				IO     int64 `json:"io"`
				CPU    int64 `json:"cpu"`
			} `json:"limits"`
			Relationships struct {
				Allocations struct {
					Data []struct {
						Attributes struct {
							IP        string `json:"ip"`
							Port      int64  `json:"port"`
							IsDefault bool   `json:"is_default"`
						} `json:"attributes"`
					} `json:"data"`
				} `json:"allocations"`
				Variables struct {
					Data []struct {
						Attributes struct {
							EnvVariable string `json:"env_variable"`
							ServerValue string `json:"server_value"`
						} `json:"attributes"`
					} `json:"data"`
				} `json:"variables"`
			} `json:"relationships"`
		} `json:"attributes"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", err.Error())
		return
	}

	a := apiResp.Attributes

	// Environment map
	envMap := make(map[string]attr.Value)
	for _, v := range a.Relationships.Variables.Data {
		envMap[v.Attributes.EnvVariable] = types.StringValue(v.Attributes.ServerValue)
	}
	environment, diags := types.MapValueFrom(ctx, types.StringType, envMap)
	resp.Diagnostics.Append(diags...)

	// Egg features (can be null)
	var eggFeatures types.List
	if a.EggFeatures == nil {
		eggFeatures = types.ListNull(types.StringType)
	} else {
		eggFeatures, diags = types.ListValueFrom(ctx, types.StringType, a.EggFeatures)
		resp.Diagnostics.Append(diags...)
	}

	// Feature limits
	featureLimitsAttrs := map[string]attr.Value{
		"databases":   types.Int64Value(a.FeatureLimits.Databases),
		"allocations": types.Int64Value(a.FeatureLimits.Allocations),
		"backups":     types.Int64Value(a.FeatureLimits.Backups),
	}
	featureLimits, diags := types.ObjectValue(
		map[string]attr.Type{
			"databases":   types.Int64Type,
			"allocations": types.Int64Type,
			"backups":     types.Int64Type,
		},
		featureLimitsAttrs,
	)
	resp.Diagnostics.Append(diags...)

	// Default allocation
	var allocIP string
	var allocPort int64
	for _, alloc := range a.Relationships.Allocations.Data {
		if alloc.Attributes.IsDefault {
			allocIP = alloc.Attributes.IP
			allocPort = alloc.Attributes.Port
			break
		}
	}

	state := serverDataModel{
		ServerID:        config.ServerID,
		ID:              types.StringValue(a.Identifier),
		Identifier:      types.StringValue(a.Identifier),
		InternalID:      types.Int64Value(a.InternalID),
		Name:            types.StringValue(a.Name),
		Description:     types.StringValue(a.Description),
		Suspended:       types.BoolValue(a.IsSuspended),
		Installing:      types.BoolValue(a.IsInstalling),
		Transferring:    types.BoolValue(a.IsTransferring),
		Node:            types.StringValue(a.Node),
		SFTPIP:          types.StringValue(a.SFTPDetails.IP),
		SFTPPort:        types.Int64Value(a.SFTPDetails.Port),
		Invocation:      types.StringValue(a.Invocation),
		DockerImage:     types.StringValue(a.DockerImage),
		Memory:          types.Int64Value(a.Limits.Memory),
		Disk:            types.Int64Value(a.Limits.Disk),
		CPU:             types.Int64Value(a.Limits.CPU),
		Swap:            types.Int64Value(a.Limits.Swap),
		IO:              types.Int64Value(a.Limits.IO),
		AllocationIP:    types.StringValue(allocIP),
		AllocationPort:  types.Int64Value(allocPort),
		Environment:     environment,
		EggFeatures:     eggFeatures,
		FeatureLimits:   featureLimits,
		UserPermissions: types.ListNull(types.StringType), // Not in response, safe to omit
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
