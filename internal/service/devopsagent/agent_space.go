// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsagent

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/devopsagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devopsagent/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_devopsagent_agent_space", name="Agent Space")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/devopsagent;devopsagent.GetAgentSpaceOutput")
// @Testing(generator=false)
func newAgentSpaceResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &agentSpaceResource{}, nil
}

type agentSpaceResource struct {
	framework.ResourceWithModel[agentSpaceResourceModel]
}

func (r *agentSpaceResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"agent_space_id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the Agent Space.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional:    true,
				Description: "The description of the Agent Space.",
			},
			names.AttrKMSKeyARN: schema.StringAttribute{
				Optional:    true,
				Description: "The ARN of the KMS key used to encrypt resources.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"locale": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The locale for the Agent Space, which determines the language used in agent responses.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required:    true,
				Description: "The name of the Agent Space.",
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
	}
}

func (r *agentSpaceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data agentSpaceResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DevOpsAgentClient(ctx)

	input := devopsagent.CreateAgentSpaceInput{
		Name: data.Name.ValueStringPointer(),
		Tags: getTagsIn(ctx),
	}

	if !data.Description.IsNull() {
		input.Description = data.Description.ValueStringPointer()
	}
	if !data.KMSKeyARN.IsNull() {
		input.KmsKeyArn = data.KMSKeyARN.ValueStringPointer()
	}
	if !data.Locale.IsNull() && !data.Locale.IsUnknown() {
		input.Locale = data.Locale.ValueStringPointer()
	}

	output, err := conn.CreateAgentSpace(ctx, &input)
	if err != nil {
		create.AddError(&response.Diagnostics, names.DevOpsAgent, create.ErrActionCreating, ResNameAgentSpace, data.Name.ValueString(), err)
		return
	}

	space := output.AgentSpace
	data.AgentSpaceID = fwflex.StringToFramework(ctx, space.AgentSpaceId)
	data.ARN = types.StringValue(agentSpaceARN(r.Meta().Partition(ctx), r.Meta().Region(ctx), r.Meta().AccountID(ctx), aws.ToString(space.AgentSpaceId)))
	data.CreatedAt = fwflex.TimeToFramework(ctx, space.CreatedAt)
	data.ID = fwflex.StringToFramework(ctx, space.AgentSpaceId)
	data.Name = fwflex.StringToFramework(ctx, space.Name)
	data.Description = fwflex.StringToFramework(ctx, space.Description)
	data.KMSKeyARN = fwflex.StringToFramework(ctx, space.KmsKeyArn)
	data.Locale = fwflex.StringToFramework(ctx, space.Locale)
	data.UpdatedAt = fwflex.TimeToFramework(ctx, space.UpdatedAt)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *agentSpaceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data agentSpaceResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DevOpsAgentClient(ctx)

	output, err := findAgentSpaceByID(ctx, conn, data.AgentSpaceID.ValueString())
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		create.AddError(&response.Diagnostics, names.DevOpsAgent, create.ErrActionReading, ResNameAgentSpace, data.AgentSpaceID.ValueString(), err)
		return
	}

	space := output.AgentSpace
	data.AgentSpaceID = fwflex.StringToFramework(ctx, space.AgentSpaceId)
	data.ARN = types.StringValue(agentSpaceARN(r.Meta().Partition(ctx), r.Meta().Region(ctx), r.Meta().AccountID(ctx), aws.ToString(space.AgentSpaceId)))
	data.CreatedAt = fwflex.TimeToFramework(ctx, space.CreatedAt)
	data.ID = fwflex.StringToFramework(ctx, space.AgentSpaceId)
	data.Name = fwflex.StringToFramework(ctx, space.Name)
	data.Description = fwflex.StringToFramework(ctx, space.Description)
	data.KMSKeyARN = fwflex.StringToFramework(ctx, space.KmsKeyArn)
	data.Locale = fwflex.StringToFramework(ctx, space.Locale)
	data.UpdatedAt = fwflex.TimeToFramework(ctx, space.UpdatedAt)

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *agentSpaceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new agentSpaceResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DevOpsAgentClient(ctx)

	if !old.Name.Equal(new.Name) || !old.Description.Equal(new.Description) || !old.Locale.Equal(new.Locale) {
		input := devopsagent.UpdateAgentSpaceInput{
			AgentSpaceId: old.AgentSpaceID.ValueStringPointer(),
		}

		if !old.Name.Equal(new.Name) {
			input.Name = new.Name.ValueStringPointer()
		}
		if !old.Description.Equal(new.Description) {
			input.Description = new.Description.ValueStringPointer()
		}
		if !old.Locale.Equal(new.Locale) {
			input.Locale = new.Locale.ValueStringPointer()
		}

		_, err := conn.UpdateAgentSpace(ctx, &input)
		if err != nil {
			create.AddError(&response.Diagnostics, names.DevOpsAgent, create.ErrActionUpdating, ResNameAgentSpace, new.AgentSpaceID.ValueString(), err)
			return
		}
	}

	// Always read back to get fresh timestamps.
	output, err := findAgentSpaceByID(ctx, conn, old.AgentSpaceID.ValueString())
	if err != nil {
		create.AddError(&response.Diagnostics, names.DevOpsAgent, create.ErrActionReading, ResNameAgentSpace, old.AgentSpaceID.ValueString(), err)
		return
	}

	space := output.AgentSpace
	new.AgentSpaceID = old.AgentSpaceID
	new.ARN = old.ARN
	new.CreatedAt = old.CreatedAt
	new.ID = old.ID
	new.Name = fwflex.StringToFramework(ctx, space.Name)
	new.Description = fwflex.StringToFramework(ctx, space.Description)
	new.Locale = fwflex.StringToFramework(ctx, space.Locale)
	new.UpdatedAt = fwflex.TimeToFramework(ctx, space.UpdatedAt)

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *agentSpaceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data agentSpaceResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DevOpsAgentClient(ctx)

	tflog.Debug(ctx, "deleting DevOps Agent Space", map[string]any{
		"agent_space_id": data.AgentSpaceID.ValueString(),
	})

	input := devopsagent.DeleteAgentSpaceInput{
		AgentSpaceId: data.AgentSpaceID.ValueStringPointer(),
	}

	_, err := conn.DeleteAgentSpace(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		create.AddError(&response.Diagnostics, names.DevOpsAgent, create.ErrActionDeleting, ResNameAgentSpace, data.AgentSpaceID.ValueString(), err)
	}
}

func (r *agentSpaceResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("agent_space_id"), request, response)
}

func findAgentSpaceByID(ctx context.Context, conn *devopsagent.Client, id string) (*devopsagent.GetAgentSpaceOutput, error) {
	input := devopsagent.GetAgentSpaceInput{
		AgentSpaceId: aws.String(id),
	}

	output, err := conn.GetAgentSpace(ctx, &input)
	if err != nil {
		return nil, fmt.Errorf("reading DevOps Agent Space (%s): %w", id, err)
	}

	if output == nil || output.AgentSpace == nil {
		return nil, fmt.Errorf("reading DevOps Agent Space (%s): empty output", id)
	}

	return output, nil
}

type agentSpaceResourceModel struct {
	framework.WithRegionModel
	AgentSpaceID types.String      `tfsdk:"agent_space_id"`
	ARN          types.String      `tfsdk:"arn"`
	CreatedAt    timetypes.RFC3339 `tfsdk:"created_at"`
	Description  types.String      `tfsdk:"description"`
	ID           types.String      `tfsdk:"id"`
	KMSKeyARN    types.String      `tfsdk:"kms_key_arn"`
	Locale       types.String      `tfsdk:"locale"`
	Name         types.String      `tfsdk:"name"`
	Tags         tftags.Map        `tfsdk:"tags"`
	TagsAll      tftags.Map        `tfsdk:"tags_all"`
	UpdatedAt    timetypes.RFC3339 `tfsdk:"updated_at"`
}

func agentSpaceARN(partition, region, accountID, agentSpaceID string) string {
	return arn.ARN{
		Partition: partition,
		Service:   "aidevops",
		Region:    region,
		AccountID: accountID,
		Resource:  "agentspace/" + agentSpaceID,
	}.String()
}
