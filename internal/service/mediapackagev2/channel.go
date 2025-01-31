// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediapackagev2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediapackagev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mediapackagev2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameChannel         = "Channel"
	ChannelFieldNamePrefix = "Channel"
)

// @FrameworkResource("aws_media_packagev2_channel", name="Channel")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/mediapackagev2;mediapackagev2.GetChannelOutput")
// @Testing(serialize=true)
// @Testing(importStateIdFunc=testAccChannelImportStateIdFunc)
// @Testing(importStateIdAttribute=name)
func newResourceChannel(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceChannel{}

	return r, nil
}

type resourceChannel struct {
	framework.ResourceWithConfigure
}

func (r *resourceChannel) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_media_packagev2_channel"
}

func (r *resourceChannel) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"channel_group_name": schema.StringAttribute{
				Required: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"input_type": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(func() []string {
						values := awstypes.InputType("").Values()
						strValues := make([]string, len(values))
						for i, v := range values {
							strValues[i] = string(v)
						}
						return strValues
					}()...),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"ingest_endpoints": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[ingestEndpointModel](ctx),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"url": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}

	response.Schema = s
}

func (r *resourceChannel) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().MediaPackageV2Client(ctx)
	var data resourceChannelData

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := mediapackagev2.CreateChannelInput{
		Tags: getTagsIn(ctx),
	}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input, fwflex.WithFieldNamePrefix(ChannelFieldNamePrefix))...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateChannel(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaPackageV2, create.ErrActionCreating, ResNameChannel, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data, fwflex.WithFieldNamePrefix(ChannelFieldNamePrefix))...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceChannel) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().MediaPackageV2Client(ctx)
	var data resourceChannelData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := findChannelByID(ctx, conn, data.ChannelGroupName.ValueString(), data.Name.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaPackageV2, create.ErrActionReading, ResNameChannel, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data, fwflex.WithFieldNamePrefix(ChannelFieldNamePrefix))...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceChannel) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().MediaPackageV2Client(ctx)
	var state, plan resourceChannelData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Calculate(ctx, plan, state)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		input := mediapackagev2.UpdateChannelInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix(ChannelFieldNamePrefix))...)
		if response.Diagnostics.HasError() {
			return
		}

		output, err := conn.UpdateChannel(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.MediaPackageV2, create.ErrActionUpdating, ResNameChannel, state.Name.String(), err),
				err.Error(),
			)
			return
		}

		response.Diagnostics.Append(fwflex.Flatten(ctx, output, &plan, fwflex.WithFieldNamePrefix(ChannelFieldNamePrefix))...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourceChannel) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().MediaPackageV2Client(ctx)
	var data resourceChannelData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting Channel", map[string]interface{}{
		names.AttrName: data.Name.ValueString(),
	})

	input := mediapackagev2.DeleteChannelInput{
		ChannelGroupName: data.ChannelGroupName.ValueStringPointer(),
		ChannelName:      data.Name.ValueStringPointer(),
	}

	_, err := conn.DeleteChannel(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaPackageV2, create.ErrActionDeleting, ResNameChannel, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	_, err = tfresource.RetryUntilNotFound(ctx, 5*time.Minute, func() (interface{}, error) {
		return findChannelByID(ctx, conn, data.ChannelGroupName.ValueString(), data.Name.ValueString())
	})

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaPackageV2, create.ErrActionWaitingForDeletion, ResNameChannel, data.Name.String(), err),
			err.Error(),
		)
	}
}

func (r *resourceChannel) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrName), request, response)
}

func (r *resourceChannel) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

type resourceChannelData struct {
	ARN              types.String `tfsdk:"arn"`
	Name             types.String `tfsdk:"name"`
	ChannelGroupName types.String `tfsdk:"channel_group_name"`
	Description      types.String `tfsdk:"description"`
	InputType        types.String `tfsdk:"input_type"`
	//IngestEndpoints  fwtypes.ListNestedObjectValueOf[ingestEndpointModel] `tfsdk:"ingest_endpoints"`
	IngestEndpoints fwtypes.SetNestedObjectValueOf[ingestEndpointModel] `tfsdk:"ingest_endpoints"`
	Tags            tftags.Map                                          `tfsdk:"tags"`
	TagsAll         tftags.Map                                          `tfsdk:"tags_all"`
}

type ingestEndpointModel struct {
	Id  types.String `tfsdk:"id"`
	Url types.String `tfsdk:"url"`
}

func findChannelByID(ctx context.Context, conn *mediapackagev2.Client, channelGroupName string, channelName string) (*mediapackagev2.GetChannelOutput, error) {
	in := &mediapackagev2.GetChannelInput{
		ChannelGroupName: aws.String(channelGroupName),
		ChannelName:      aws.String(channelName),
	}

	out, err := conn.GetChannel(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastRequest: in,
			LastError:   err,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
