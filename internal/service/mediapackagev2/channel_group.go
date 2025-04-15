// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediapackagev2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediapackagev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mediapackagev2/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameChannelGroup         = "Channel Group"
	ChannelGroupFieldNamePrefix = "ChannelGroup"
)

// @FrameworkResource("aws_media_packagev2_channel_group", name="Channel Group")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/mediapackagev2;mediapackagev2.GetChannelGroupOutput")
// @Testing(serialize=true)
// @Testing(importStateIdFunc=testAccChannelGroupImportStateIdFunc)
// @Testing(importStateIdAttribute=name)
func newResourceChannelGroup(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceChannelGroup{}

	return r, nil
}

type resourceChannelGroup struct {
	framework.ResourceWithConfigure
}

func (r *resourceChannelGroup) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"egress_domain": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}

	response.Schema = s
}

func (r *resourceChannelGroup) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().MediaPackageV2Client(ctx)
	var data resourceChannelGroupData

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := mediapackagev2.CreateChannelGroupInput{
		Tags: getTagsIn(ctx),
	}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input, fwflex.WithFieldNamePrefix(ChannelGroupFieldNamePrefix))...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateChannelGroup(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaPackageV2, create.ErrActionCreating, ResNameChannelGroup, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data, fwflex.WithFieldNamePrefix(ChannelGroupFieldNamePrefix))...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceChannelGroup) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().MediaPackageV2Client(ctx)
	var data resourceChannelGroupData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := findChannelGroupByID(ctx, conn, data.Name.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaPackageV2, create.ErrActionReading, ResNameChannelGroup, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data, fwflex.WithFieldNamePrefix(ChannelGroupFieldNamePrefix))...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceChannelGroup) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().MediaPackageV2Client(ctx)
	var state, plan resourceChannelGroupData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		input := mediapackagev2.UpdateChannelGroupInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix(ChannelGroupFieldNamePrefix))...)
		if response.Diagnostics.HasError() {
			return
		}

		output, err := conn.UpdateChannelGroup(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.MediaPackageV2, create.ErrActionUpdating, ResNameChannelGroup, state.Name.String(), err),
				err.Error(),
			)
			return
		}

		response.Diagnostics.Append(fwflex.Flatten(ctx, output, &plan, fwflex.WithFieldNamePrefix(ChannelGroupFieldNamePrefix))...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourceChannelGroup) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().MediaPackageV2Client(ctx)
	var data resourceChannelGroupData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting Channel Group", map[string]any{
		names.AttrName: data.Name.ValueString(),
	})

	input := mediapackagev2.DeleteChannelGroupInput{
		ChannelGroupName: data.Name.ValueStringPointer(),
	}

	_, err := conn.DeleteChannelGroup(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaPackageV2, create.ErrActionDeleting, ResNameChannelGroup, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	_, err = tfresource.RetryUntilNotFound(ctx, 5*time.Minute, func() (any, error) {
		return findChannelGroupByID(ctx, conn, data.Name.ValueString())
	})

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaPackageV2, create.ErrActionWaitingForDeletion, ResNameChannelGroup, data.Name.String(), err),
			err.Error(),
		)
	}
}

func (r *resourceChannelGroup) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrName), request, response)
}

type resourceChannelGroupData struct {
	ARN          types.String `tfsdk:"arn"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	EgressDomain types.String `tfsdk:"egress_domain"`
	Tags         tftags.Map   `tfsdk:"tags"`
	TagsAll      tftags.Map   `tfsdk:"tags_all"`
}

func findChannelGroupByID(ctx context.Context, conn *mediapackagev2.Client, id string) (*mediapackagev2.GetChannelGroupOutput, error) {
	in := &mediapackagev2.GetChannelGroupInput{
		ChannelGroupName: aws.String(id),
	}

	out, err := conn.GetChannelGroup(ctx, in)

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
