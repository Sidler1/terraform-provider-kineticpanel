package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ServerReinstallResource{}

// ServerReinstallResource triggers a server reinstall (wipe + redeploy).
type ServerReinstallResource struct {
	client *Client
}

// reinstallModel holds the resource state.
type reinstallModel struct {
	ServerID types.String `tfsdk:"server_id"`
	Force    types.Bool   `tfsdk:"force"` // bypass confirmation if supported
	ID       types.String `tfsdk:"id"`    // synthetic: "<server_id>-reinstall"
}

func NewServerReinstallResource() resource.Resource {
	return &ServerReinstallResource{}
}

func (r *ServerReinstallResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_reinstall"
}

func (r *ServerReinstallResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reinstalls a Kinetic Panel server (wipes data and redeploys).",
		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Short server identifier (e.g. `abc123`).",
			},
			"force": schema.BoolAttribute{
				Optional:    true,
				Description: "Bypass confirmation if the panel supports it. Default: false.",
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Synthetic resource ID (`<server_id>-reinstall`).",
			},
		},
	}
}

func (r *ServerReinstallResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ServerReinstallResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan reinstallModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := "/servers/" + plan.ServerID.ValueString() + "/settings/reinstall"
	payload := map[string]bool{}
	if !plan.Force.IsNull() {
		payload["force"] = plan.Force.ValueBool()
	}

	_, err := r.client.Post(path, payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to trigger reinstall", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.ServerID.ValueString() + "-reinstall")
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ServerReinstallResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state reinstallModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// No read-back — reinstall is a one-time action
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ServerReinstallResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan reinstallModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Reinstall again if force changes or re-applied
	path := "/servers/" + plan.ServerID.ValueString() + "/settings/reinstall"
	payload := map[string]bool{}
	if !plan.Force.IsNull() {
		payload["force"] = plan.Force.ValueBool()
	}

	_, err := r.client.Post(path, payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update reinstall", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.ServerID.ValueString() + "-reinstall")
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ServerReinstallResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No-op: reinstall cannot be undone
	resp.State.RemoveResource(ctx)
}
