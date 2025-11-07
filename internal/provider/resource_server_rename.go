package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ServerRenameResource{}

// ServerRenameResource updates a server's name and description.
type ServerRenameResource struct {
	client *Client
}

// renameModel holds the resource state.
type renameModel struct {
	ServerID    types.String `tfsdk:"server_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ID          types.String `tfsdk:"id"` // synthetic: "<server_id>-rename"
}

func NewServerRenameResource() resource.Resource {
	return &ServerRenameResource{}
}

func (r *ServerRenameResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_rename"
}

func (r *ServerRenameResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("server_id"), req, resp)
}

func (r *ServerRenameResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Renames a Kinetic Panel server and updates its description (Client API).",
		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Short server identifier (e.g. `abc123`).",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "New server name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "New server description.",
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Synthetic resource ID (`<server_id>-rename`).",
			},
		},
	}
}

func (r *ServerRenameResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ServerRenameResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan renameModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pth := "/servers/" + plan.ServerID.ValueString() + "/settings/rename"
	payload := map[string]string{
		"name": plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() {
		payload["description"] = plan.Description.ValueString()
	}

	_, err := r.client.Post(pth, payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to rename server", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.ServerID.ValueString() + "-rename")
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ServerRenameResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state renameModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Read just returns stored values — actual name is in data_server
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ServerRenameResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan renameModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pth := "/servers/" + plan.ServerID.ValueString() + "/settings/rename"
	payload := map[string]string{
		"name": plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() {
		payload["description"] = plan.Description.ValueString()
	}

	_, err := r.client.Post(pth, payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update server name", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.ServerID.ValueString() + "-rename")
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ServerRenameResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No-op: renaming is not reversible via API
	resp.State.RemoveResource(ctx)
}
