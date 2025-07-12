// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_fsx_s3_access_point_attachment", name="S3 Access Point Attachment")
func newS3AccessPointAttachmentResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &s3AccessPointAttachmentResource{}

	r.SetDefaultCreateTimeout(15 * time.Minute)
	r.SetDefaultDeleteTimeout(15 * time.Minute)

	return r, nil
}

type s3AccessPointAttachmentResource struct {
	framework.ResourceWithModel[s3AccessPointAttachmentResourceModel]
	framework.WithTimeouts
	framework.WithNoUpdate
}

func (r *s3AccessPointAttachmentResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z0-9][a-z0-9-]{1,48}[a-z0-9]$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.S3AccessPointAttachmentType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *s3AccessPointAttachmentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data s3AccessPointAttachmentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().FSxClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input fsx.CreateAndAttachS3AccessPointInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientRequestToken = aws.String(sdkid.UniqueId())

	_, err := conn.CreateAndAttachS3AccessPoint(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating FSx S3 Access Point Attachment (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.

	// TODO Wait.

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *s3AccessPointAttachmentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data s3AccessPointAttachmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().FSxClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	output, err := findS3AccessPointAttachmentByName(ctx, conn, name)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading FSx S3 Access Point Attachment (%s)", name), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *s3AccessPointAttachmentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data s3AccessPointAttachmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().FSxClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	input := fsx.DetachAndDeleteS3AccessPointInput{
		ClientRequestToken: aws.String(sdkid.UniqueId()),
		Name:               aws.String(name),
	}

	_, err := conn.DetachAndDeleteS3AccessPoint(ctx, &input)

	if errs.IsA[*awstypes.S3AccessPointAttachmentNotFound](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting FSx S3 Access Point Attachment (%s)", name), err.Error())

		return
	}

	// TODO Wait.
}

func (r *s3AccessPointAttachmentResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrName), request, response)
}

func findS3AccessPointAttachmentByName(ctx context.Context, conn *fsx.Client, name string) (*awstypes.S3AccessPointAttachment, error) {
	input := fsx.DescribeS3AccessPointAttachmentsInput{
		Names: []string{name},
	}

	return findS3AccessPointAttachment(ctx, conn, &input, tfslices.PredicateTrue[*awstypes.S3AccessPointAttachment]())
}

func findS3AccessPointAttachment(ctx context.Context, conn *fsx.Client, input *fsx.DescribeS3AccessPointAttachmentsInput, filter tfslices.Predicate[*awstypes.S3AccessPointAttachment]) (*awstypes.S3AccessPointAttachment, error) {
	output, err := findS3AccessPointAttachments(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findS3AccessPointAttachments(ctx context.Context, conn *fsx.Client, input *fsx.DescribeS3AccessPointAttachmentsInput, filter tfslices.Predicate[*awstypes.S3AccessPointAttachment]) ([]awstypes.S3AccessPointAttachment, error) {
	var output []awstypes.S3AccessPointAttachment

	pages := fsx.NewDescribeS3AccessPointAttachmentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.S3AccessPointAttachmentNotFound](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.S3AccessPointAttachments {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusS3AccessPointAttachment(ctx context.Context, conn *fsx.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findS3AccessPointAttachmentByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Lifecycle), nil
	}
}

type s3AccessPointAttachmentResourceModel struct {
	framework.WithRegionModel
	Name types.String                                             `tfsdk:"name"`
	Type fwtypes.StringEnum[awstypes.S3AccessPointAttachmentType] `tfsdk:"type"`
}
