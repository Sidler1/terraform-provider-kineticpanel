package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ServerActivityLogsDataSource{}

// ServerActivityLogsDataSource fetches recent console logs for a server.
type ServerActivityLogsDataSource struct {
	client *Client
}

// logsModel holds the data source state.
type logsModel struct {
	ServerID   types.String `tfsdk:"server_id"`
	Lines      types.Int64  `tfsdk:"lines"`      // how many lines to fetch (default 50)
	Logs       types.List   `tfsdk:"logs"`       // list of log lines
	Timestamps types.List   `tfsdk:"timestamps"` // list of timestamps (ISO 8601)
}

func NewServerActivityLogsDataSource() datasource.DataSource {
	return &ServerActivityLogsDataSource{}
}

func (d *ServerActivityLogsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_activity_logs"
}

func (d *ServerActivityLogsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches recent console activity logs for a Kinetic Panel server (Client API).",
		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				Required:    true,
				Description: "Short server identifier (e.g. `abc123`).",
			},
			"lines": schema.Int64Attribute{
				Optional:    true,
				Description: "Number of recent log lines to fetch. Default: 50. Max: 100.",
			},
			"logs": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "List of console log lines (most recent first).",
			},
			"timestamps": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "List of timestamps (ISO 8601) corresponding to each log line.",
			},
		},
	}
}

func (d *ServerActivityLogsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServerActivityLogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config struct {
		ServerID types.String `tfsdk:"server_id"`
		Lines    types.Int64  `tfsdk:"lines"`
	}
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverID := config.ServerID.ValueString()
	lines := int(config.Lines.ValueInt64())
	if lines == 0 {
		lines = 50
	}
	if lines > 100 {
		lines = 100
	}

	// Build URL with query param: ?logs=50
	u, _ := url.Parse("/servers/" + serverID + "/websocket")
	q := u.Query()
	q.Set("logs", fmt.Sprintf("%d", lines))
	u.RawQuery = q.Encode()
	path := u.String()

	body, err := d.client.Get(path)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to fetch logs for server %s: %v", serverID, err))
		return
	}

	var apiResp struct {
		Data []struct {
			Event     string   `json:"event"`
			Args      []string `json:"args"`
			Timestamp string   `json:"timestamp"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", err.Error())
		return
	}

	var logLines []string
	var timestamps []string

	for _, entry := range apiResp.Data {
		if entry.Event == "console output" && len(entry.Args) > 0 {
			// Clean up ANSI codes and trim
			line := strings.TrimSpace(entry.Args[0])
			line = stripANSI(line)
			if line != "" {
				logLines = append(logLines, line)
				timestamps = append(timestamps, entry.Timestamp)
			}
		}
	}

	// Reverse to most recent first
	for i := len(logLines)/2 - 1; i >= 0; i-- {
		opp := len(logLines) - 1 - i
		logLines[i], logLines[opp] = logLines[opp], logLines[i]
		timestamps[i], timestamps[opp] = timestamps[opp], timestamps[i]
	}

	logsList, diags := types.ListValueFrom(ctx, types.StringType, logLines)
	resp.Diagnostics.Append(diags...)
	tsList, diags := types.ListValueFrom(ctx, types.StringType, timestamps)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := logsModel{
		ServerID:   config.ServerID,
		Lines:      types.Int64Value(int64(lines)),
		Logs:       logsList,
		Timestamps: tsList,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// stripANSI removes ANSI color codes (basic)
func stripANSI(str string) string {
	const ansi = "\x1b\\[[0-9;]*[a-zA-Z]"
	re := regexp.MustCompile(ansi)
	return re.ReplaceAllString(str, "")
}
