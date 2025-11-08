package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ServersDataSource{}

type ServersDataSource struct{ client *Client }

type serversDataModel struct {
	Servers types.List `tfsdk:"servers"`
}

type serverItemModel struct {
	ID              types.String `tfsdk:"id"`
	Identifier      types.String `tfsdk:"identifier"`
	InternalID      types.Int64  `tfsdk:"internal_id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	Suspended       types.Bool   `tfsdk:"is_suspended"`
	Installing      types.Bool   `tfsdk:"is_installing"`
	Transferring    types.Bool   `tfsdk:"is_transferring"`
	Node            types.String `tfsdk:"node"`
	SFTPIP          types.String `tfsdk:"sftp_ip"`
	SFTPPort        types.Int64  `tfsdk:"sftp_port"`
	Invocation      types.String `tfsdk:"invocation"`
	DockerImage     types.String `tfsdk:"docker_image"`
	Memory          types.Int64  `tfsdk:"memory"`
	Disk            types.Int64  `tfsdk:"disk"`
	CPU             types.Int64  `tfsdk:"cpu"`
	Swap            types.Int64  `tfsdk:"swap"`
	IO              types.Int64  `tfsdk:"io"`
	AllocationIP    types.String `tfsdk:"allocation_ip"`
	AllocationPort  types.Int64  `tfsdk:"allocation_port"`
	Environment     types.Map    `tfsdk:"environment"`
	EggFeatures     types.List   `tfsdk:"egg_features"`
	FeatureLimits   types.Object `tfsdk:"feature_limits"`
	UserPermissions types.List   `tfsdk:"user_permissions"`
}

func NewServersDataSource() datasource.DataSource { return &ServersDataSource{} }

func (d *ServersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_servers"
}

func (d *ServersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all servers accessible via Client API.",
		Attributes: map[string]schema.Attribute{
			"servers": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":              schema.StringAttribute{Computed: true},
						"identifier":      schema.StringAttribute{Computed: true},
						"internal_id":     schema.Int64Attribute{Computed: true},
						"name":            schema.StringAttribute{Computed: true},
						"description":     schema.StringAttribute{Computed: true},
						"is_suspended":    schema.BoolAttribute{Computed: true},
						"is_installing":   schema.BoolAttribute{Computed: true},
						"is_transferring": schema.BoolAttribute{Computed: true},
						"node":            schema.StringAttribute{Computed: true},
						"sftp_ip":         schema.StringAttribute{Computed: true},
						"sftp_port":       schema.Int64Attribute{Computed: true},
						"invocation":      schema.StringAttribute{Computed: true},
						"docker_image":    schema.StringAttribute{Computed: true},
						"memory":          schema.Int64Attribute{Computed: true},
						"disk":            schema.Int64Attribute{Computed: true},
						"cpu":             schema.Int64Attribute{Computed: true},
						"swap":            schema.Int64Attribute{Computed: true},
						"io":              schema.Int64Attribute{Computed: true},
						"allocation_ip":   schema.StringAttribute{Computed: true},
						"allocation_port": schema.Int64Attribute{Computed: true},
						"environment":     schema.MapAttribute{ElementType: types.StringType, Computed: true},
						"egg_features":    schema.ListAttribute{ElementType: types.StringType, Computed: true},
						"feature_limits": schema.ObjectAttribute{
							AttributeTypes: map[string]attr.Type{
								"databases":   types.Int64Type,
								"allocations": types.Int64Type,
								"backups":     types.Int64Type,
							},
							Computed: true,
						},
						"user_permissions": schema.ListAttribute{ElementType: types.StringType, Computed: true},
					},
				},
			},
		},
	}
}

func (d *ServersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", "")
		return
	}
	d.client = client
}

func (d *ServersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state serversDataModel
	body, err := d.client.Get("")
	if err != nil {
		resp.Diagnostics.AddError("API Error", err.Error())
		return
	}

	var apiResp struct {
		Data []struct {
			Attributes struct {
				Identifier     string `json:"identifier"`
				InternalID     int64  `json:"internal_id"`
				Name           string `json:"name"`
				Description    string `json:"description"`
				IsSuspended    bool   `json:"is_suspended"`
				IsInstalling   bool   `json:"is_installing"`
				IsTransferring bool   `json:"is_transferring"`
				Node           string `json:"node"`
				SFTPDetails    struct {
					IP   string `json:"ip"`
					Port int64  `json:"port"`
				} `json:"sftp_details"`
				Invocation    string   `json:"invocation"`
				DockerImage   string   `json:"docker_image"`
				EggFeatures   []string `json:"egg_features"`
				FeatureLimits struct {
					Databases   int64 `json:"databases"`
					Allocations int64 `json:"allocations"`
					Backups     int64 `json:"backups"`
				} `json:"feature_limits"`
				UserPermissions []string `json:"user_permissions"`
				Limits          struct {
					Memory int64 `json:"memory"`
					Swap   int64 `json:"swap"`
					Disk   int64 `json:"disk"`
					IO     int64 `json:"io"`
					CPU    int64 `json:"cpu"`
				} `json:"limits"`
				Relationships struct {
					Allocations struct {
						Data []struct {
							Attributes struct {
								ID        int64  `json:"id"`
								IP        string `json:"ip"`
								Port      int64  `json:"port"`
								IsDefault bool   `json:"is_default"`
							} `json:"attributes"`
						} `json:"data"`
					} `json:"allocations"`
					Variables struct {
						Data []struct {
							Attributes struct {
								EnvVariable string `json:"env_variable"`
								ServerValue string `json:"server_value"`
							} `json:"attributes"`
						} `json:"data"`
					} `json:"variables"`
				} `json:"relationships"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", err.Error())
		return
	}

	var servers []attr.Value
	for _, srv := range apiResp.Data {
		a := srv.Attributes
		envMap := make(map[string]attr.Value)
		for _, v := range a.Relationships.Variables.Data {
			envMap[v.Attributes.EnvVariable] = types.StringValue(v.Attributes.ServerValue)
		}
		env, _ := types.MapValue(types.StringType, envMap)

		eggFeatures, _ := types.ListValueFrom(ctx, types.StringType, a.EggFeatures)
		userPerms, _ := types.ListValueFrom(ctx, types.StringType, a.UserPermissions)

		var allocIP string
		var allocPort int64
		for _, alloc := range a.Relationships.Allocations.Data {
			if alloc.Attributes.IsDefault {
				allocIP = alloc.Attributes.IP
				allocPort = alloc.Attributes.Port
				break
			}
		}

		limits := map[string]attr.Value{
			"databases":   types.Int64Value(a.FeatureLimits.Databases),
			"allocations": types.Int64Value(a.FeatureLimits.Allocations),
			"backups":     types.Int64Value(a.FeatureLimits.Backups),
		}
		featureLimits, _ := types.ObjectValue(map[string]attr.Type{
			"databases":   types.Int64Type,
			"allocations": types.Int64Type,
			"backups":     types.Int64Type,
		}, limits)

		item := serverItemModel{
			ID:              types.StringValue(a.Identifier),
			Identifier:      types.StringValue(a.Identifier),
			InternalID:      types.Int64Value(a.InternalID),
			Name:            types.StringValue(a.Name),
			Description:     types.StringValue(a.Description),
			Suspended:       types.BoolValue(a.IsSuspended),
			Installing:      types.BoolValue(a.IsInstalling),
			Transferring:    types.BoolValue(a.IsTransferring),
			Node:            types.StringValue(a.Node),
			SFTPIP:          types.StringValue(a.SFTPDetails.IP),
			SFTPPort:        types.Int64Value(a.SFTPDetails.Port),
			Invocation:      types.StringValue(a.Invocation),
			DockerImage:     types.StringValue(a.DockerImage),
			Memory:          types.Int64Value(a.Limits.Memory),
			Disk:            types.Int64Value(a.Limits.Disk),
			CPU:             types.Int64Value(a.Limits.CPU),
			Swap:            types.Int64Value(a.Limits.Swap),
			IO:              types.Int64Value(a.Limits.IO),
			AllocationIP:    types.StringValue(allocIP),
			AllocationPort:  types.Int64Value(allocPort),
			Environment:     env,
			EggFeatures:     eggFeatures,
			FeatureLimits:   featureLimits,
			UserPermissions: userPerms,
		}

		obj, _ := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"id": types.StringType, "identifier": types.StringType, "internal_id": types.Int64Type,
			"name": types.StringType, "description": types.StringType, "is_suspended": types.BoolType,
			"is_installing": types.BoolType, "is_transferring": types.BoolType, "node": types.StringType,
			"sftp_ip": types.StringType, "sftp_port": types.Int64Type, "invocation": types.StringType,
			"docker_image": types.StringType, "memory": types.Int64Type, "disk": types.Int64Type,
			"cpu": types.Int64Type, "swap": types.Int64Type, "io": types.Int64Type,
			"allocation_ip": types.StringType, "allocation_port": types.Int64Type,
			"environment":  types.MapType{ElemType: types.StringType},
			"egg_features": types.ListType{ElemType: types.StringType},
			"feature_limits": types.ObjectType{AttrTypes: map[string]attr.Type{
				"databases": types.Int64Type, "allocations": types.Int64Type, "backups": types.Int64Type,
			}},
			"user_permissions": types.ListType{ElemType: types.StringType},
		}, item)

		servers = append(servers, obj)
	}

	serversList, _ := types.ListValue(types.ObjectType{AttrTypes: map[string]attr.Type{
		"id": types.StringType, "identifier": types.StringType, "internal_id": types.Int64Type,
		"name": types.StringType, "description": types.StringType, "is_suspended": types.BoolType,
		"is_installing": types.BoolType, "is_transferring": types.BoolType, "node": types.StringType,
		"sftp_ip": types.StringType, "sftp_port": types.Int64Type, "invocation": types.StringType,
		"docker_image": types.StringType, "memory": types.Int64Type, "disk": types.Int64Type,
		"cpu": types.Int64Type, "swap": types.Int64Type, "io": types.Int64Type,
		"allocation_ip": types.StringType, "allocation_port": types.Int64Type,
		"environment":  types.MapType{ElemType: types.StringType},
		"egg_features": types.ListType{ElemType: types.StringType},
		"feature_limits": types.ObjectType{AttrTypes: map[string]attr.Type{
			"databases": types.Int64Type, "allocations": types.Int64Type, "backups": types.Int64Type,
		}},
		"user_permissions": types.ListType{ElemType: types.StringType},
	}}, servers)

	state.Servers = serversList
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
