// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package devopsguru

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/devopsguru"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devopsguru/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_devopsguru_event_sources_config", name="Event Sources Config")
// @SingletonIdentity(identityDuplicateAttributes="id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/devopsguru;devopsguru.DescribeEventSourcesConfigOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(generator=false)
// @Testing(preIdentityVersion="v5.100.0")
func newEventSourcesConfigResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &eventSourcesConfigResource{}, nil
}

const (
	ResNameEventSourcesConfig = "Event Sources Config"
)

type eventSourcesConfigResource struct {
	framework.ResourceWithModel[eventSourcesConfigResourceModel]
	framework.WithImportByIdentity
}

func (r *eventSourcesConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrRegion)),
		},
		Blocks: map[string]schema.Block{
			"event_sources": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[eventSourcesData](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"amazon_code_guru_profiler": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[amazonCodeGuruProfilerData](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrStatus: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.EventSourceOptInStatus](),
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *eventSourcesConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DevOpsGuruClient(ctx)

	var plan eventSourcesConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(r.Meta().Region(ctx))

	in := &devopsguru.UpdateEventSourcesConfigInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, &plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.UpdateEventSourcesConfig(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DevOpsGuru, create.ErrActionCreating, ResNameEventSourcesConfig, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *eventSourcesConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DevOpsGuruClient(ctx)

	var state eventSourcesConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findEventSourcesConfig(ctx, conn)
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DevOpsGuru, create.ErrActionSetting, ResNameEventSourcesConfig, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *eventSourcesConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update is a no-op
}

func (r *eventSourcesConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DevOpsGuruClient(ctx)

	var state eventSourcesConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &devopsguru.UpdateEventSourcesConfigInput{
		EventSources: &awstypes.EventSourcesConfig{
			AmazonCodeGuruProfiler: &awstypes.AmazonCodeGuruProfilerIntegration{
				Status: awstypes.EventSourceOptInStatusDisabled,
			},
		},
	}

	_, err := conn.UpdateEventSourcesConfig(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DevOpsGuru, create.ErrActionDeleting, ResNameEventSourcesConfig, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func findEventSourcesConfig(ctx context.Context, conn *devopsguru.Client) (*devopsguru.DescribeEventSourcesConfigOutput, error) {
	in := &devopsguru.DescribeEventSourcesConfigInput{}

	out, err := conn.DescribeEventSourcesConfig(ctx, in)
	if err != nil {
		return nil, err
	}

	if out == nil || out.EventSources == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

type eventSourcesConfigResourceModel struct {
	framework.WithRegionModel
	EventSources fwtypes.ListNestedObjectValueOf[eventSourcesData] `tfsdk:"event_sources"`
	ID           types.String                                      `tfsdk:"id"`
}

type eventSourcesData struct {
	AmazonCodeGuruProfiler fwtypes.ListNestedObjectValueOf[amazonCodeGuruProfilerData] `tfsdk:"amazon_code_guru_profiler"`
}

type amazonCodeGuruProfilerData struct {
	Status fwtypes.StringEnum[awstypes.EventSourceOptInStatus] `tfsdk:"status"`
}
