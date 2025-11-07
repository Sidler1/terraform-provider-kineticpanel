package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ServerResource{}

// serverAPIResponse matches the exact JSON structure returned by KineticPanel API
type serverAPIResponse struct {
	Object     string `json:"object"`
	Attributes struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		User        int64  `json:"user"`
		Egg         int64  `json:"egg"`
		Location    int64  `json:"location"`
		Node        int64  `json:"node"`
		Memory      int64  `json:"memory"`
		Disk        int64  `json:"disk"`
		CPU         int64  `json:"cpu"`
		DockerImage string `json:"docker_image"`
		Startup     string `json:"startup"`
	} `json:"attributes"`
}

type ServerResource struct {
	client *Client
}

type serverModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	UserID      types.Int64  `tfsdk:"user_id"`
	EggID       types.Int64  `tfsdk:"egg_id"`
	LocationID  types.Int64  `tfsdk:"location_id"`
	NodeID      types.Int64  `tfsdk:"node_id"`
	Memory      types.Int64  `tfsdk:"memory"`
	Disk        types.Int64  `tfsdk:"disk"`
	CPU         types.Int64  `tfsdk:"cpu"`
	DockerImage types.String `tfsdk:"docker_image"`
	StartupCmd  types.String `tfsdk:"startup_command"`
}

func NewServerResource() resource.Resource {
	return &ServerResource{}
}

func (r *ServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *ServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a game server on Kinetic Panel using the Application API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"user_id": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"egg_id": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"location_id": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"node_id": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"memory": schema.Int64Attribute{
				Required: true,
			},
			"disk": schema.Int64Attribute{
				Required: true,
			},
			"cpu": schema.Int64Attribute{
				Required: true,
			},
			"docker_image": schema.StringAttribute{
				Required: true,
			},
			"startup_command": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (r *ServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = client
}

// helper: convert model → API payload
func modelToPayload(plan serverModel) map[string]any {
	return map[string]any{
		"name":         plan.Name.ValueString(),
		"user":         plan.UserID.ValueInt64(),
		"egg":          plan.EggID.ValueInt64(),
		"location":     plan.LocationID.ValueInt64(),
		"node":         plan.NodeID.ValueInt64(),
		"memory":       plan.Memory.ValueInt64(),
		"disk":         plan.Disk.ValueInt64(),
		"cpu":          plan.CPU.ValueInt64(),
		"docker_image": plan.DockerImage.ValueString(),
		"startup":      plan.StartupCmd.ValueString(),
	}
}

// helper: API response → model
func apiToModel(apiResp serverAPIResponse) serverModel {
	a := apiResp.Attributes
	return serverModel{
		ID:          types.Int64Value(a.ID),
		Name:        types.StringValue(a.Name),
		UserID:      types.Int64Value(a.User),
		EggID:       types.Int64Value(a.Egg),
		LocationID:  types.Int64Value(a.Location),
		NodeID:      types.Int64Value(a.Node),
		Memory:      types.Int64Value(a.Memory),
		Disk:        types.Int64Value(a.Disk),
		CPU:         types.Int64Value(a.CPU),
		DockerImage: types.StringValue(a.DockerImage),
		StartupCmd:  types.StringValue(a.Startup),
	}
}

func (r *ServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serverModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.Post("/servers", modelToPayload(plan))
	if err != nil {
		resp.Diagnostics.AddError("API Create Error", err.Error())
		return
	}

	var apiResp serverAPIResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", err.Error())
		return
	}

	state := apiToModel(apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serverModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.Get("/servers/" + strconv.FormatInt(state.ID.ValueInt64(), 10))
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Read Error", err.Error())
		return
	}

	var apiResp serverAPIResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", err.Error())
		return
	}

	state = apiToModel(apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serverModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Patch("/servers/"+strconv.FormatInt(plan.ID.ValueInt64(), 10), modelToPayload(plan))
	if err != nil {
		resp.Diagnostics.AddError("API Update Error", err.Error())
		return
	}

	// Refresh state from API after update
	readReq := resource.ReadRequest{
		State: req.State, // use pre-update state (contains known ID)
	}
	var readResp resource.ReadResponse
	r.Read(ctx, readReq, &readResp)

	resp.Diagnostics.Append(readResp.Diagnostics...)
	resp.State = readResp.State
}

func (r *ServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serverModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete("/servers/" + strconv.FormatInt(state.ID.ValueInt64(), 10))
	if err != nil && !strings.Contains(err.Error(), "404") {
		resp.Diagnostics.AddError("API Delete Error", err.Error())
	}
}
