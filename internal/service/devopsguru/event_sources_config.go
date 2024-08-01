// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Event Sources Config")
func newResourceEventSourcesConfig(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceEventSourcesConfig{}, nil
}

const (
	ResNameEventSourcesConfig = "Event Sources Config"
)

type resourceEventSourcesConfig struct {
	framework.ResourceWithConfigure
}

func (r *resourceEventSourcesConfig) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_devopsguru_event_sources_config"
}

func (r *resourceEventSourcesConfig) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
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

func (r *resourceEventSourcesConfig) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DevOpsGuruClient(ctx)

	var plan resourceEventSourcesConfigData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(r.Meta().Region)

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

func (r *resourceEventSourcesConfig) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DevOpsGuruClient(ctx)

	var state resourceEventSourcesConfigData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findEventSourcesConfig(ctx, conn)
	if tfresource.NotFound(err) {
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

func (r *resourceEventSourcesConfig) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update is a no-op
}

func (r *resourceEventSourcesConfig) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DevOpsGuruClient(ctx)

	var state resourceEventSourcesConfigData
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

func (r *resourceEventSourcesConfig) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func findEventSourcesConfig(ctx context.Context, conn *devopsguru.Client) (*devopsguru.DescribeEventSourcesConfigOutput, error) {
	in := &devopsguru.DescribeEventSourcesConfigInput{}

	out, err := conn.DescribeEventSourcesConfig(ctx, in)
	if err != nil {
		return nil, err
	}

	if out == nil || out.EventSources == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceEventSourcesConfigData struct {
	EventSources fwtypes.ListNestedObjectValueOf[eventSourcesData] `tfsdk:"event_sources"`
	ID           types.String                                      `tfsdk:"id"`
}

type eventSourcesData struct {
	AmazonCodeGuruProfiler fwtypes.ListNestedObjectValueOf[amazonCodeGuruProfilerData] `tfsdk:"amazon_code_guru_profiler"`
}

type amazonCodeGuruProfilerData struct {
	Status fwtypes.StringEnum[awstypes.EventSourceOptInStatus] `tfsdk:"status"`
}
