// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3vectors

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3vectors"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3vectors/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource("aws_s3vectors_vector_bucket", name="VectorBucket")
// @Tags
func newVectorBucketResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &vectorBucketResource{}

	return r, nil
}

type vectorBucketResource struct {
	framework.ResourceWithModel[vectorBucketResourceModel]
	framework.WithNoUpdate
}

func (r *vectorBucketResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"vector_bucket_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vector_bucket_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *vectorBucketResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data vectorBucketResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3VectorsClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.VectorBucketName)
	var input s3vectors.CreateVectorBucketInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.CreateVectorBucket(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Vectors Vector Bucket (%s)", name), err.Error())

		return
	}
}

func (r *vectorBucketResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data vectorBucketResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3VectorsClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.VectorBucketARN)
	output, err := findVectorBuckeyByARN(ctx, conn, arn)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Vectors Vector Bucket (%s)", arn), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
}

func (r *vectorBucketResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data vectorBucketResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3VectorsClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.VectorBucketARN)
	input := s3vectors.DeleteVectorBucketInput{
		VectorBucketArn: aws.String(arn),
	}
	_, err := conn.DeleteVectorBucket(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Vectors Vector Bucket (%s)", arn), err.Error())

		return
	}
}

func findVectorBuckeyByARN(ctx context.Context, conn *s3vectors.Client, arn string) (*awstypes.VectorBucket, error) {
	input := s3vectors.GetVectorBucketInput{
		VectorBucketArn: aws.String(arn),
	}
	output, err := conn.GetVectorBucket(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.VectorBucket == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.VectorBucket, nil
}

type vectorBucketResourceModel struct {
	framework.WithRegionModel
	VectorBucketARN  types.String `tfsdk:"vector_bucket_arn"`
	VectorBucketName types.String `tfsdk:"vector_bucket_name"`
}
