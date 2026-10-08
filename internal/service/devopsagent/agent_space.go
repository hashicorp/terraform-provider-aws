// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsagent

import (
	"context"
	"errors"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/devopsagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devopsagent/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_devopsagent_agent_space", name="Agent Space")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/devopsagent;devopsagent.GetAgentSpaceOutput")
// @Testing(importStateIdAttribute="agent_space_id")
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
				Description: "Unique identifier of the Agent Space.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional:    true,
				Description: "Description of the Agent Space.",
			},
			names.AttrKMSKeyARN: schema.StringAttribute{
				Optional:    true,
				Description: "ARN of the KMS key used to encrypt resources.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"locale": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Locale for the Agent Space, which determines the language used in agent responses.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required:    true,
				Description: "Name of the Agent Space.",
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
	conn := r.Meta().DevOpsAgentClient(ctx)

	var data agentSpaceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	var input devopsagent.CreateAgentSpaceInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateAgentSpace(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.Name.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output.AgentSpace, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *agentSpaceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().DevOpsAgentClient(ctx)

	var data agentSpaceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	output, err := findAgentSpaceByID(ctx, conn, data.AgentSpaceID.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.AgentSpaceID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output.AgentSpace, &data))
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, output.Tags)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *agentSpaceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().DevOpsAgentClient(ctx)

	var old, new agentSpaceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	if response.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input devopsagent.UpdateAgentSpaceInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input, diff.IgnoredFieldNamesOpts()...))
		if response.Diagnostics.HasError() {
			return
		}
		input.AgentSpaceId = old.AgentSpaceID.ValueStringPointer()

		output, err := conn.UpdateAgentSpace(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, old.AgentSpaceID.ValueString())
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output.AgentSpace, &new))
		if response.Diagnostics.HasError() {
			return
		}
	} else {
		// Tags only update does not change the updated at in upstream
		new.UpdatedAt = old.UpdatedAt
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *agentSpaceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().DevOpsAgentClient(ctx)

	var data agentSpaceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

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
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.AgentSpaceID.ValueString())
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
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil || output.AgentSpace == nil {
		return nil, smarterr.NewError(errors.New("empty output"))
	}

	return output, nil
}

type agentSpaceResourceModel struct {
	framework.WithRegionModel
	AgentSpaceID types.String      `tfsdk:"agent_space_id"`
	ARN          types.String      `tfsdk:"arn"`
	CreatedAt    timetypes.RFC3339 `tfsdk:"created_at"`
	Description  types.String      `tfsdk:"description"`
	KMSKeyARN    types.String      `tfsdk:"kms_key_arn"`
	Locale       types.String      `tfsdk:"locale"`
	Name         types.String      `tfsdk:"name"`
	Tags         tftags.Map        `tfsdk:"tags"`
	TagsAll      tftags.Map        `tfsdk:"tags_all"`
	UpdatedAt    timetypes.RFC3339 `tfsdk:"updated_at"`
}

func (r *agentSpaceResource) flatten(ctx context.Context, space *awstypes.AgentSpace, data *agentSpaceResourceModel) diag.Diagnostics {
	diags := fwflex.Flatten(ctx, space, data)
	if diags.HasError() {
		return diags
	}
	// The API does not return an ARN.
	data.ARN = types.StringValue(r.Meta().RegionalARN(ctx, "aidevops", "agentspace/"+data.AgentSpaceID.ValueString()))

	return diags
}
