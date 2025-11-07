package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ServerStartupVariableResource{}

// ServerStartupVariableResource updates a single startup environment variable.
type ServerStartupVariableResource struct {
	client *Client
}

// variableModel holds the resource state.
type variableModel struct {
	ServerID types.String `tfsdk:"server_id"`
	Key      types.String `tfsdk:"key"`   // e.g. "MEMORYSIZE"
	Value    types.String `tfsdk:"value"` // e.g. "2048"
	ID       types.String `tfsdk:"id"`    // synthetic: "<server_id>-var-<key>"
}

func NewServerStartupVariableResource() resource.Resource {
	return &ServerStartupVariableResource{}
}

func (r *ServerStartupVariableResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_startup_variable"
}

func (r *ServerStartupVariableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID", "Expected format: <server_id>:<key>")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("server_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), parts[1])...)
}

func (r *ServerStartupVariableResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Updates a single startup environment variable for a Kinetic Panel server (Client API).",
		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Short server identifier (e.g. `abc123`).",
			},
			"key": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Environment variable key (e.g. `MEMORYSIZE`, `SERVER_JAR`).",
			},
			"value": schema.StringAttribute{
				Required:    true,
				Description: "Value to set for the variable.",
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Synthetic resource ID (`<server_id>-var-<key>`).",
			},
		},
	}
}

func (r *ServerStartupVariableResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ServerStartupVariableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan variableModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pth := "/servers/" + plan.ServerID.ValueString() + "/startup/variable"
	payload := map[string]string{
		"key":   plan.Key.ValueString(),
		"value": plan.Value.ValueString(),
	}

	_, err := r.client.Post(pth, payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update startup variable", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.ServerID.ValueString() + "-var-" + plan.Key.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ServerStartupVariableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state variableModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// No read-back — use data_server_startup to verify
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ServerStartupVariableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan variableModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pth := "/servers/" + plan.ServerID.ValueString() + "/startup/variable"
	payload := map[string]string{
		"key":   plan.Key.ValueString(),
		"value": plan.Value.ValueString(),
	}

	_, err := r.client.Post(pth, payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update startup variable", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.ServerID.ValueString() + "-var-" + plan.Key.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ServerStartupVariableResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No-op: cannot delete variables via API
	resp.State.RemoveResource(ctx)
}
