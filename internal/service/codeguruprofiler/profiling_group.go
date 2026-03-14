// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package codeguruprofiler

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codeguruprofiler"
	awstypes "github.com/aws/aws-sdk-go-v2/service/codeguruprofiler/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_codeguruprofiler_profiling_group", name="Profiling Group")
// @Tags(identifierAttribute="arn")
func newProfilingGroupResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &profilingGroupResource{}

	return r, nil
}

const (
	ResNameProfilingGroup = "Profiling Group"
)

type profilingGroupResource struct {
	framework.ResourceWithModel[profilingGroupResourceModel]
	framework.WithImportByID
}

func (r *profilingGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	computePlatform := fwtypes.StringEnumType[awstypes.ComputePlatform]()

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"compute_platform": schema.StringAttribute{
				CustomType: computePlatform,
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"agent_orchestration_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[agentOrchestrationConfig](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"profiling_enabled": schema.BoolAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}

func (r *profilingGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CodeGuruProfilerClient(ctx)

	var plan profilingGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &codeguruprofiler.CreateProfilingGroupInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)

	if resp.Diagnostics.HasError() {
		return
	}

	in.ProfilingGroupName = flex.StringFromFramework(ctx, plan.Name)
	in.ClientToken = aws.String(sdkid.UniqueId())
	in.Tags = getTagsIn(ctx)

	out, err := conn.CreateProfilingGroup(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CodeGuruProfiler, create.ErrActionCreating, ResNameProfilingGroup, plan.Name.ValueString(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.ProfilingGroup == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CodeGuruProfiler, create.ErrActionCreating, ResNameProfilingGroup, plan.Name.ValueString(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	state := plan

	resp.Diagnostics.Append(flex.Flatten(ctx, out.ProfilingGroup, &state)...)

	state.ID = flex.StringToFramework(ctx, out.ProfilingGroup.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *profilingGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CodeGuruProfilerClient(ctx)

	var state profilingGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findProfilingGroupByName(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CodeGuruProfiler, create.ErrActionSetting, ResNameProfilingGroup, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, out.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *profilingGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().CodeGuruProfilerClient(ctx)

	var plan, state profilingGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.AgentOrchestrationConfig.Equal(state.AgentOrchestrationConfig) {
		in := &codeguruprofiler.UpdateProfilingGroupInput{}
		resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)

		if resp.Diagnostics.HasError() {
			return
		}

		in.ProfilingGroupName = flex.StringFromFramework(ctx, state.ID)
		out, err := conn.UpdateProfilingGroup(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CodeGuruProfiler, create.ErrActionUpdating, ResNameProfilingGroup, plan.ID.String(), err),
				err.Error(),
			)
			return
		}

		if out == nil || out.ProfilingGroup == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CodeGuruProfiler, create.ErrActionUpdating, ResNameProfilingGroup, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *profilingGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CodeGuruProfilerClient(ctx)

	var state profilingGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &codeguruprofiler.DeleteProfilingGroupInput{
		ProfilingGroupName: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteProfilingGroup(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CodeGuruProfiler, create.ErrActionDeleting, ResNameProfilingGroup, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func findProfilingGroupByName(ctx context.Context, conn *codeguruprofiler.Client, name string) (*awstypes.ProfilingGroupDescription, error) {
	in := &codeguruprofiler.DescribeProfilingGroupInput{
		ProfilingGroupName: aws.String(name),
	}

	out, err := conn.DescribeProfilingGroup(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.ProfilingGroup == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out.ProfilingGroup, nil
}

type profilingGroupResourceModel struct {
	framework.WithRegionModel
	ARN                      types.String                                              `tfsdk:"arn"`
	AgentOrchestrationConfig fwtypes.ListNestedObjectValueOf[agentOrchestrationConfig] `tfsdk:"agent_orchestration_config"`
	ComputePlatform          fwtypes.StringEnum[awstypes.ComputePlatform]              `tfsdk:"compute_platform"`
	ID                       types.String                                              `tfsdk:"id"`
	Name                     types.String                                              `tfsdk:"name"`
	Tags                     tftags.Map                                                `tfsdk:"tags"`
	TagsAll                  tftags.Map                                                `tfsdk:"tags_all"`
}

type agentOrchestrationConfig struct {
	ProfilingEnabled types.Bool `tfsdk:"profiling_enabled"`
}
