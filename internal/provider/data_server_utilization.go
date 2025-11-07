package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ServerUtilizationDataSource{}

// ServerUtilizationDataSource fetches real-time resource usage of a server.
type ServerUtilizationDataSource struct {
	client *Client
}

// utilizationModel holds the data source state.
type utilizationModel struct {
	ServerID  types.String  `tfsdk:"server_id"`
	State     types.String  `tfsdk:"state"`        // running, offline, etc.
	CPU       types.Float64 `tfsdk:"cpu_percent"`  // % of allocated CPU
	Memory    types.Int64   `tfsdk:"memory_bytes"` // current usage
	MemoryMB  types.Float64 `tfsdk:"memory_mb"`    // computed for convenience
	Disk      types.Int64   `tfsdk:"disk_bytes"`
	DiskMB    types.Float64 `tfsdk:"disk_mb"`
	NetworkRX types.Int64   `tfsdk:"network_rx_bytes"`
	NetworkTX types.Int64   `tfsdk:"network_tx_bytes"`
	Uptime    types.Int64   `tfsdk:"uptime_seconds"`
}

func NewServerUtilizationDataSource() datasource.DataSource {
	return &ServerUtilizationDataSource{}
}

func (d *ServerUtilizationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_utilization"
}

func (d *ServerUtilizationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches real-time resource utilization for a Kinetic Panel server (Client API).",
		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				Required:    true,
				Description: "Short server identifier (e.g. `abc123`).",
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "Current server state: `running`, `starting`, `stopping`, `offline`.",
			},
			"cpu_percent": schema.Float64Attribute{
				Computed:    true,
				Description: "Current CPU usage as percentage of allocated limit.",
			},
			"memory_bytes": schema.Int64Attribute{
				Computed:    true,
				Description: "Current memory usage in bytes.",
			},
			"memory_mb": schema.Float64Attribute{
				Computed:    true,
				Description: "Current memory usage in MB (rounded to 2 decimals).",
			},
			"disk_bytes": schema.Int64Attribute{
				Computed:    true,
				Description: "Current disk usage in bytes.",
			},
			"disk_mb": schema.Float64Attribute{
				Computed:    true,
				Description: "Current disk usage in MB (rounded to 2 decimals).",
			},
			"network_rx_bytes": schema.Int64Attribute{
				Computed:    true,
				Description: "Total received network bytes since boot.",
			},
			"network_tx_bytes": schema.Int64Attribute{
				Computed:    true,
				Description: "Total transmitted network bytes since boot.",
			},
			"uptime_seconds": schema.Int64Attribute{
				Computed:    true,
				Description: "Server uptime in seconds.",
			},
		},
	}
}

func (d *ServerUtilizationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServerUtilizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config struct {
		ServerID types.String `tfsdk:"server_id"`
	}
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverID := config.ServerID.ValueString()
	path := "/servers/" + serverID + "/utilization"

	body, err := d.client.Get(path)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to fetch utilization for server %s: %v", serverID, err))
		return
	}

	var apiResp struct {
		State   string `json:"state"`
		Memory  int64  `json:"memory"`
		CPU     int64  `json:"cpu"`
		Disk    int64  `json:"disk"`
		Network struct {
			RX int64 `json:"rx"`
			TX int64 `json:"tx"`
		} `json:"network"`
		Uptime int64 `json:"uptime"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", err.Error())
		return
	}

	// Convert bytes to MB with 2 decimal precision
	memoryMB := float64(apiResp.Memory) / (1024 * 1024)
	diskMB := float64(apiResp.Disk) / (1024 * 1024)

	state := utilizationModel{
		ServerID:  config.ServerID,
		State:     types.StringValue(apiResp.State),
		CPU:       types.Float64Value(float64(apiResp.CPU)),
		Memory:    types.Int64Value(apiResp.Memory),
		MemoryMB:  types.Float64Value(roundTo(memoryMB, 2)),
		Disk:      types.Int64Value(apiResp.Disk),
		DiskMB:    types.Float64Value(roundTo(diskMB, 2)),
		NetworkRX: types.Int64Value(apiResp.Network.RX),
		NetworkTX: types.Int64Value(apiResp.Network.TX),
		Uptime:    types.Int64Value(apiResp.Uptime),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// roundTo rounds float64 to given decimal places
func roundTo(n float64, decimals uint) float64 {
	factor := 1.0
	for i := uint(0); i < decimals; i++ {
		factor *= 10
	}
	return float64(int64(n*factor+0.5)) / factor
}
