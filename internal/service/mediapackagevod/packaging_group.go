// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediapackagevod

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediapackagevod"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mediapackagevod/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNamePackagingGroup         = "Packaging Group"
	PackagingGroupFieldNamePrefix = "PackagingGroup"
)

// @FrameworkResource("aws_mediapackagevod_packaging_group", name="Packaging Group")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/mediapackagevod;mediapackagevod.DescribePackagingGroupOutput")
// @Testing(serialize=true)
// @Testing(importStateIdFunc=testAccPackagingGroupImportStateIdFunc)
// @Testing(importStateIdAttribute=name)
func newResourcePackagingGroup(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePackagingGroup{}

	return r, nil
}

type resourcePackagingGroup struct {
	framework.ResourceWithConfigure
}

func (r *resourcePackagingGroup) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_mediapackagevod_packaging_group"
}

func (r *resourcePackagingGroup) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"authorization": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[authorizationModel](ctx),
				Optional:   true,
				Computed:   true,
				AttributeTypes: map[string]attr.Type{
					"cdn_identifier_secret": types.StringType,
					"secrets_role_arn":      types.StringType,
				},
			},
			"egress_access_logs": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[egressAccessLogsModel](ctx),
				Optional:   true,
				Computed:   true,
				AttributeTypes: map[string]attr.Type{
					"log_group_name": types.StringType,
				},
			},
			"domain_name": schema.StringAttribute{
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

func (r *resourcePackagingGroup) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().MediaPackageVODClient(ctx)
	var data resourcePackagingGroupData

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := mediapackagevod.CreatePackagingGroupInput{
		Tags: getTagsIn(ctx),
	}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input, fwflex.WithFieldNamePrefix(PackagingGroupFieldNamePrefix))...)
	if response.Diagnostics.HasError() {
		return
	}

	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 1*time.Minute, func() (interface{}, error) {
		return conn.CreatePackagingGroup(ctx, &input)
	}, "UnprocessableEntityException")

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaPackageVOD, create.ErrActionCreating, ResNamePackagingGroup, data.Id.String(), err),
			err.Error(),
		)
		return
	}

	output := outputRaw.(*mediapackagevod.CreatePackagingGroupOutput)
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data, fwflex.WithFieldNamePrefix(PackagingGroupFieldNamePrefix))...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourcePackagingGroup) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().MediaPackageVODClient(ctx)
	var data resourcePackagingGroupData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := findPackagingGroupByID(ctx, conn, data.Id.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaPackageVOD, create.ErrActionReading, ResNamePackagingGroup, data.Id.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data, fwflex.WithFieldNamePrefix(PackagingGroupFieldNamePrefix))...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourcePackagingGroup) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().MediaPackageVODClient(ctx)
	var state, plan resourcePackagingGroupData

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
		input := mediapackagevod.UpdatePackagingGroupInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix(PackagingGroupFieldNamePrefix))...)
		if response.Diagnostics.HasError() {
			return
		}

		outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 1*time.Minute, func() (interface{}, error) {
			return conn.UpdatePackagingGroup(ctx, &input)
		}, "UnprocessableEntityException")

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.MediaPackageVOD, create.ErrActionUpdating, ResNamePackagingGroup, state.Id.String(), err),
				err.Error(),
			)
			return
		}

		output := outputRaw.(*mediapackagevod.UpdatePackagingGroupOutput)
		response.Diagnostics.Append(fwflex.Flatten(ctx, output, &plan, fwflex.WithFieldNamePrefix(PackagingGroupFieldNamePrefix))...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourcePackagingGroup) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().MediaPackageVODClient(ctx)
	var data resourcePackagingGroupData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting Packaging Group", map[string]interface{}{
		names.AttrName: data.Id.ValueString(),
	})

	input := mediapackagevod.DeletePackagingGroupInput{
		Id: data.Id.ValueStringPointer(),
	}

	_, err := conn.DeletePackagingGroup(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaPackageVOD, create.ErrActionDeleting, ResNamePackagingGroup, data.Id.String(), err),
			err.Error(),
		)
		return
	}

	_, err = tfresource.RetryUntilNotFound(ctx, 5*time.Minute, func() (interface{}, error) {
		return findPackagingGroupByID(ctx, conn, data.Id.ValueString())
	})

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaPackageVOD, create.ErrActionWaitingForDeletion, ResNamePackagingGroup, data.Id.String(), err),
			err.Error(),
		)
	}
}

func (r *resourcePackagingGroup) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrName), request, response)
}

func (r *resourcePackagingGroup) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

type resourcePackagingGroupData struct {
	ARN              types.String                                 `tfsdk:"arn"`
	Id               types.String                                 `tfsdk:"name"`
	Authorization    fwtypes.ObjectValueOf[authorizationModel]    `tfsdk:"authorization"`
	EgressAccessLogs fwtypes.ObjectValueOf[egressAccessLogsModel] `tfsdk:"egress_access_logs"`
	DomainName       types.String                                 `tfsdk:"domain_name"`
	Tags             tftags.Map                                   `tfsdk:"tags"`
	TagsAll          tftags.Map                                   `tfsdk:"tags_all"`
}

type egressAccessLogsModel struct {
	LogGroupName types.String `tfsdk:"log_group_name"`
}

type authorizationModel struct {
	CdnIdentifierSecret types.String `tfsdk:"cdn_identifier_secret"`
	SecretsRoleArn      types.String `tfsdk:"secrets_role_arn"`
}

func findPackagingGroupByID(ctx context.Context, conn *mediapackagevod.Client, name string) (*mediapackagevod.DescribePackagingGroupOutput, error) {
	in := &mediapackagevod.DescribePackagingGroupInput{
		Id: aws.String(name),
	}

	out, err := conn.DescribePackagingGroup(ctx, in)

	if errs.IsA[*awstypes.NotFoundException](err) {
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
