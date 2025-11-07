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

var _ resource.Resource = &ServerDockerImageResource{}

// ServerDockerImageResource updates the Docker image for a server.
type ServerDockerImageResource struct {
	client *Client
}

// dockerImageModel holds the resource state.
type dockerImageModel struct {
	ServerID    types.String `tfsdk:"server_id"`
	DockerImage types.String `tfsdk:"docker_image"`
	ID          types.String `tfsdk:"id"` // synthetic: "<server_id>-docker"
}

func NewServerDockerImageResource() resource.Resource {
	return &ServerDockerImageResource{}
}

func (r *ServerDockerImageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_docker_image"
}

func (r *ServerDockerImageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("server_id"), req, resp)
}

func (r *ServerDockerImageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Updates the Docker image used by a Kinetic Panel server (Client API).",
		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Short server identifier (e.g. `abc123`).",
			},
			"docker_image": schema.StringAttribute{
				Required:    true,
				Description: "Docker image tag (e.g. `ghcr.io/pterodactyl/wings:latest`).",
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Synthetic resource ID (`<server_id>-docker`).",
			},
		},
	}
}

func (r *ServerDockerImageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ServerDockerImageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dockerImageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pth := "/servers/" + plan.ServerID.ValueString() + "/settings/docker-image"
	payload := map[string]string{
		"docker_image": plan.DockerImage.ValueString(),
	}

	_, err := r.client.Post(pth, payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update Docker image", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.ServerID.ValueString() + "-docker")
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ServerDockerImageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dockerImageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// No read-back — use data_server_startup to verify
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ServerDockerImageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dockerImageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pth := "/servers/" + plan.ServerID.ValueString() + "/settings/docker-image"
	payload := map[string]string{
		"docker_image": plan.DockerImage.ValueString(),
	}

	_, err := r.client.Post(pth, payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update Docker image", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.ServerID.ValueString() + "-docker")
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ServerDockerImageResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No-op: Docker image change is not reversible via API
	resp.State.RemoveResource(ctx)
}
