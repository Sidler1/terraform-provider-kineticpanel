package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ServerResource{}

type ServerResource struct {
	client *Client
}

func NewServerResource() resource.Resource {
	return &ServerResource{}
}

func (r *ServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *ServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a game server on Kinetic Panel (Pterodactyl).",
		Attributes: map[string]schema.Attribute{
			// Minimal required fields for POST /api/application/servers
			"id": schema.Int64Attribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"user_id": schema.Int64Attribute{
				Required: true,
			},
			"egg_id": schema.Int64Attribute{
				Required: true,
			},
			// add more: nest_id, location_id, memory, disk, cpu, etc.
		},
	}
}
