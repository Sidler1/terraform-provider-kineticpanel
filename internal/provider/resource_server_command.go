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

var _ resource.Resource = &ServerCommandResource{}

// ServerCommandResource sends a console command to a Kinetic Panel server (Client API).
type ServerCommandResource struct {
	client *Client
}

// serverCommandModel holds the Terraform state for this resource.
type serverCommandModel struct {
	ServerID types.String `tfsdk:"server_id"` // short identifier, e.g. "1a2b3c"
	Command  types.String `tfsdk:"command"`   // console command to run
	ID       types.String `tfsdk:"id"`        // synthetic ID (server_id + "-cmd")
}

// NewServerCommandResource returns a new instance of the resource.
func NewServerCommandResource() resource.Resource {
	return &ServerCommandResource{}
}

// Metadata sets the Terraform type name.
func (r *ServerCommandResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_command"
}

// Schema defines the resource attributes.
func (r *ServerCommandResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Sends a console command to a Kinetic Panel server (Client API).",
		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Server identifier (short ID, e.g. `1a2b3c`).",
			},
			"command": schema.StringAttribute{
				Required:    true,
				Description: "Console command to execute.",
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Synthetic resource ID (`<server_id>-cmd`).",
			},
		},
	}
}

// Configure injects the HTTP client that was built in the provider.
func (r *ServerCommandResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create sends the command (first time the resource is applied).
func (r *ServerCommandResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serverCommandModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := "/servers/" + plan.ServerID.ValueString() + "/command"
	payload := map[string]string{"command": plan.Command.ValueString()}

	_, err := r.client.Post(path, payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to send command", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.ServerID.ValueString() + "-cmd")
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read simply returns the stored state – the API does not return a command object.
func (r *ServerCommandResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serverCommandModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update re-sends the command when the `command` attribute changes.
func (r *ServerCommandResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serverCommandModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := "/servers/" + plan.ServerID.ValueString() + "/command"
	payload := map[string]string{"command": plan.Command.ValueString()}

	_, err := r.client.Post(path, payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update command", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.ServerID.ValueString() + "-cmd")
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete is a no-op – the command has already been executed.
func (r *ServerCommandResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}
