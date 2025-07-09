// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudfront_anycast_ip_list", name="Anycast IP List")
// @Tags(identifierAttribute="arn")
func newAnycastIPListResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &anycastIPListResource{}

	return r, nil
}

type anycastIPListResource struct {
	framework.ResourceWithModel[anycastIPListResourceModel]
	framework.WithImportByID
}

func (r *anycastIPListResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"anycast_ips": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrARN: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"etag": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ip_count": schema.Int32Attribute{
				Required: true,
				Validators: []validator.Int32{
					int32validator.OneOf(3, 21),
				},
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9-_]{1,64}$`), "must contain between 1 and 64 lowercase letters, numbers, hyphens, or underscores"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *anycastIPListResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data anycastIPListResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input cloudfront.CreateAnycastIpListInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	if tags := getTagsIn(ctx); len(tags) > 0 {
		input.Tags = &awstypes.Tags{
			Items: tags,
		}
	}

	output, err := conn.CreateAnycastIpList(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating CloudFront Anycast IP List (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	anycastIPList := output.AnycastIpList
	data.AnycastIPs = fwflex.FlattenFrameworkStringValueListOfString(ctx, anycastIPList.AnycastIps)
	data.ARN = fwflex.StringToFramework(ctx, anycastIPList.Arn)
	data.ETag = fwflex.StringToFramework(ctx, output.ETag)
	data.ID = fwflex.StringToFramework(ctx, anycastIPList.Id)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *anycastIPListResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data anycastIPListResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.ID)
	output, err := findAnycastIPListByID(ctx, conn, id)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudFront Anycast IP List (%s)", id), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *anycastIPListResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data anycastIPListResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.ID)
	input := cloudfront.DeleteAnycastIpListInput{
		Id:      aws.String(id),
		IfMatch: fwflex.StringFromFramework(ctx, data.ETag),
	}
	_, err := conn.DeleteAnycastIpList(ctx, &input)

	if errs.IsA[*awstypes.EntityNotFound](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting CloudFront Anycast IP List (%s)", id), err.Error())

		return
	}
}

func findAnycastIPListByID(ctx context.Context, conn *cloudfront.Client, id string) (*awstypes.AnycastIpList, error) {
	input := cloudfront.GetAnycastIpListInput{
		Id: aws.String(id),
	}

	return findAnycastIPList(ctx, conn, &input)
}

func findAnycastIPList(ctx context.Context, conn *cloudfront.Client, input *cloudfront.GetAnycastIpListInput) (*awstypes.AnycastIpList, error) {
	output, err := conn.GetAnycastIpList(ctx, input)

	if errs.IsA[*awstypes.EntityNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AnycastIpList == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AnycastIpList, nil
}

type anycastIPListResourceModel struct {
	AnycastIPs fwtypes.ListOfString `tfsdk:"anycast_ips"`
	ARN        types.String         `tfsdk:"arn"`
	ETag       types.String         `tfsdk:"etag"`
	ID         types.String         `tfsdk:"id"`
	IPCount    types.Int32          `tfsdk:"ip_count"`
	Name       types.String         `tfsdk:"name"`
	Tags       tftags.Map           `tfsdk:"tags"`
	TagsAll    tftags.Map           `tfsdk:"tags_all"`
}
