// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3vectors_vector_bucket_policy", name="Vector Bucket Policy")
// @ArnIdentity("vector_bucket_arn")
// @Testing(importIgnore="policy")
// @Testing(hasNoPreExistingResource=true)
func newVectorBucketPolicyResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &vectorBucketPolicyResource{}

	return r, nil
}

type vectorBucketPolicyResource struct {
	framework.ResourceWithModel[vectorBucketPolicyResourceModel]
	framework.WithImportByIdentity
}

func (r *vectorBucketPolicyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrPolicy: schema.StringAttribute{
				CustomType: fwtypes.IAMPolicyType,
				Required:   true,
			},
			"vector_bucket_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *vectorBucketPolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data vectorBucketPolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3VectorsClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.VectorBucketARN)
	var input s3vectors.PutVectorBucketPolicyInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutVectorBucketPolicy(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Vectors Vector Bucket Policy (%s)", arn), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *vectorBucketPolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data vectorBucketPolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3VectorsClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.VectorBucketARN)
	output, err := findVectorBucketPolicyByARN(ctx, conn, arn)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Vectors Vector Bucket Policy (%s)", arn), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *vectorBucketPolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new vectorBucketPolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3VectorsClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, new.VectorBucketARN)
	var input s3vectors.PutVectorBucketPolicyInput
	response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutVectorBucketPolicy(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating S3 Vectors Vector Bucket Policy (%s)", arn), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *vectorBucketPolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data vectorBucketPolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3VectorsClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.VectorBucketARN)
	input := s3vectors.DeleteVectorBucketPolicyInput{
		VectorBucketArn: aws.String(arn),
	}
	_, err := conn.DeleteVectorBucketPolicy(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Vectors Vector Bucket Policy (%s)", arn), err.Error())

		return
	}
}

func findVectorBucketPolicyByARN(ctx context.Context, conn *s3vectors.Client, arn string) (*s3vectors.GetVectorBucketPolicyOutput, error) {
	input := s3vectors.GetVectorBucketPolicyInput{
		VectorBucketArn: aws.String(arn),
	}

	return findVectorBucketPolicy(ctx, conn, &input)
}

func findVectorBucketPolicy(ctx context.Context, conn *s3vectors.Client, input *s3vectors.GetVectorBucketPolicyInput) (*s3vectors.GetVectorBucketPolicyOutput, error) {
	output, err := conn.GetVectorBucketPolicy(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Policy == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type vectorBucketPolicyResourceModel struct {
	framework.WithRegionModel
	Policy          fwtypes.IAMPolicy `tfsdk:"policy"`
	VectorBucketARN fwtypes.ARN       `tfsdk:"vector_bucket_arn"`
}
