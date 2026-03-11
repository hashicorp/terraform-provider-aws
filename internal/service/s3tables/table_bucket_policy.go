// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package s3tables

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3tables/types"
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
)

// @FrameworkResource("aws_s3tables_table_bucket_policy", name="Table Bucket Policy")
// @ArnIdentity("table_bucket_arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3tables;s3tables.GetTableBucketPolicyOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(preIdentityVersion="6.19.0")
// @Testing(importIgnore="resource_policy")
func newTableBucketPolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &tableBucketPolicyResource{}, nil
}

type tableBucketPolicyResource struct {
	framework.ResourceWithModel[tableBucketPolicyResourceModel]
	framework.WithImportByIdentity
}

func (r *tableBucketPolicyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_policy": schema.StringAttribute{
				CustomType: fwtypes.IAMPolicyType,
				Required:   true,
			},
			"table_bucket_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *tableBucketPolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data tableBucketPolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	tableBucketARN := fwflex.StringValueFromFramework(ctx, data.TableBucketARN)
	var input s3tables.PutTableBucketPolicyInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutTableBucketPolicy(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Tables Table Bucket Policy (%s)", tableBucketARN), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *tableBucketPolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data tableBucketPolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	tableBucketARN := fwflex.StringValueFromFramework(ctx, data.TableBucketARN)
	output, err := findTableBucketPolicyByARN(ctx, conn, tableBucketARN)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table Bucket Policy (%s)", tableBucketARN), err.Error())

		return
	}

	data.ResourcePolicy = fwtypes.IAMPolicyValue(aws.ToString(output.ResourcePolicy))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *tableBucketPolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new tableBucketPolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	tableBucketARN := fwflex.StringValueFromFramework(ctx, new.TableBucketARN)
	var input s3tables.PutTableBucketPolicyInput
	response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutTableBucketPolicy(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating S3 Tables Table Bucket Policy (%s)", tableBucketARN), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, new)...)
}

func (r *tableBucketPolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data tableBucketPolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	tableBucketARN := fwflex.StringValueFromFramework(ctx, data.TableBucketARN)
	input := s3tables.DeleteTableBucketPolicyInput{
		TableBucketARN: aws.String(tableBucketARN),
	}
	_, err := conn.DeleteTableBucketPolicy(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Tables Table Bucket Policy (%s)", tableBucketARN), err.Error())

		return
	}
}

func findTableBucketPolicyByARN(ctx context.Context, conn *s3tables.Client, tableBucketARN string) (*s3tables.GetTableBucketPolicyOutput, error) {
	input := s3tables.GetTableBucketPolicyInput{
		TableBucketARN: aws.String(tableBucketARN),
	}

	return findTableBucketPolicy(ctx, conn, &input)
}

func findTableBucketPolicy(ctx context.Context, conn *s3tables.Client, input *s3tables.GetTableBucketPolicyInput) (*s3tables.GetTableBucketPolicyOutput, error) {
	output, err := conn.GetTableBucketPolicy(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || aws.ToString(output.ResourcePolicy) == "" {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type tableBucketPolicyResourceModel struct {
	framework.WithRegionModel
	ResourcePolicy fwtypes.IAMPolicy `tfsdk:"resource_policy"`
	TableBucketARN fwtypes.ARN       `tfsdk:"table_bucket_arn"`
}
