// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource("aws_s3tables_table_bucket_replication", name="Table Bucket Replication")
// @ArnIdentity("table_bucket_arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3tables;s3tables.GetTableBucketReplicationOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
func newTableBucketReplicationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &tableBucketReplicationResource{}, nil
}

type tableBucketReplicationResource struct {
	framework.ResourceWithModel[tableBucketReplicationResourceModel]
	framework.WithImportByIdentity
}

func (r *tableBucketReplicationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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

func (r *tableBucketReplicationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data tableBucketReplicationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	tableBucketARN := fwflex.StringValueFromFramework(ctx, data.TableBucketARN)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *tableBucketReplicationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data tableBucketReplicationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	tableBucketARN := fwflex.StringValueFromFramework(ctx, data.TableBucketARN)
	output, err := findTableBucketReplicationByARN(ctx, conn, tableBucketARN)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table Bucket Replication (%s)", tableBucketARN), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *tableBucketReplicationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new tableBucketReplicationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	tableBucketARN := fwflex.StringValueFromFramework(ctx, new.TableBucketARN)

	response.Diagnostics.Append(response.State.Set(ctx, new)...)
}

func (r *tableBucketReplicationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data tableBucketReplicationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	tableBucketARN := fwflex.StringValueFromFramework(ctx, data.TableBucketARN)
}

func findTableBucketReplicationByARN(ctx context.Context, conn *s3tables.Client, tableBucketARN string) (*s3tables.GetTableBucketPolicyOutput, error) {
	input := s3tables.GetTableBucketPolicyInput{
		TableBucketARN: aws.String(tableBucketARN),
	}

	return findTableBucketPolicy(ctx, conn, &input)
}

func findTableBucketReplication(ctx context.Context, conn *s3tables.Client, input *s3tables.GetTableBucketPolicyInput) (*s3tables.GetTableBucketPolicyOutput, error) {
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
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type tableBucketReplicationResourceModel struct {
	framework.WithRegionModel
	TableBucketARN fwtypes.ARN `tfsdk:"table_bucket_arn"`
}
