package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ServerPowerResource{}

type ServerPowerResource struct {
	client *Client
}

type serverPowerModel struct {
	ServerID types.String `tfsdk:"server_id"`
	Signal   types.String `tfsdk:"signal"`
	ID       types.String `tfsdk:"id"`
}

func NewServerPowerResource() resource.Resource {
	return &ServerPowerResource{}
}

func (r *ServerPowerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_power"
}

func (r *ServerPowerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("server_id"), req, resp)
}

func (r *ServerPowerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Sends a power signal to a Kinetic Panel server (Client API).",
		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Server identifier (short ID, e.g. 1a2b3c4d).",
			},
			"signal": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("start", "stop", "restart", "kill"),
				},
				Description: "Power action to perform.",
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ServerPowerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", "Expected *Client")
		return
	}
	r.client = client
}

func (r *ServerPowerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serverPowerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pth := "/servers/" + plan.ServerID.ValueString() + "/power"
	_, err := r.client.Post(pth, map[string]string{"signal": plan.Signal.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Failed to send power signal", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.ServerID.ValueString() + "-power")
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ServerPowerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serverPowerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ServerPowerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serverPowerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pth := "/servers/" + plan.ServerID.ValueString() + "/power"
	_, err := r.client.Post(pth, map[string]string{"signal": plan.Signal.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update power signal", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.ServerID.ValueString() + "-power")
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ServerPowerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No-op
}
