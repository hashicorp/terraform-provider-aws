// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package s3tables

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	tableBucketMetricsConfigurationPropagationTimeout = 2 * time.Minute
)

// @FrameworkResource("aws_s3tables_table_bucket_metrics_configuration", name="Table Bucket Metrics Configuration")
// @ArnIdentity("table_bucket_arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3tables;s3tables.GetTableBucketMetricsConfigurationOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
// @Testing(plannableImportAction="NoOp")
func newTableBucketMetricsConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &tableBucketMetricsConfigurationResource{}, nil
}

type tableBucketMetricsConfigurationResource struct {
	framework.ResourceWithModel[tableBucketMetricsConfigurationResourceModel]
	framework.WithImportByIdentity
}

func (r *tableBucketMetricsConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"metrics_configuration_id": schema.StringAttribute{
				// The API does not document a stable format for the metrics configuration ID,
				// and all CRUD operations address the configuration by table bucket ARN.
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

func (r *tableBucketMetricsConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data tableBucketMetricsConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	tableBucketARN := fwflex.StringValueFromFramework(ctx, data.TableBucketARN)
	input := s3tables.PutTableBucketMetricsConfigurationInput{
		TableBucketARN: aws.String(tableBucketARN),
	}
	_, err := tfresource.RetryWhenIsOneOf2[*s3tables.PutTableBucketMetricsConfigurationOutput, *awstypes.ConflictException, *awstypes.NotFoundException](ctx, tableBucketMetricsConfigurationPropagationTimeout, func(ctx context.Context) (*s3tables.PutTableBucketMetricsConfigurationOutput, error) {
		return conn.PutTableBucketMetricsConfiguration(ctx, &input)
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Tables Table Bucket Metrics Configuration (%s)", tableBucketARN), err.Error())

		return
	}

	output, err := tfresource.RetryWhenNotFound(ctx, tableBucketMetricsConfigurationPropagationTimeout, func(ctx context.Context) (*s3tables.GetTableBucketMetricsConfigurationOutput, error) {
		return findTableBucketMetricsConfigurationByARN(ctx, conn, tableBucketARN)
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table Bucket Metrics Configuration (%s)", tableBucketARN), err.Error())

		return
	}

	data.MetricsConfigurationID = fwflex.StringToFramework(ctx, output.Id)
	data.TableBucketARN = fwtypes.ARNValue(aws.ToString(output.TableBucketARN))

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *tableBucketMetricsConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data tableBucketMetricsConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	tableBucketARN := fwflex.StringValueFromFramework(ctx, data.TableBucketARN)
	output, err := findTableBucketMetricsConfigurationByARN(ctx, conn, tableBucketARN)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table Bucket Metrics Configuration (%s)", tableBucketARN), err.Error())

		return
	}

	data.MetricsConfigurationID = fwflex.StringToFramework(ctx, output.Id)
	data.TableBucketARN = fwtypes.ARNValue(aws.ToString(output.TableBucketARN))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *tableBucketMetricsConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data tableBucketMetricsConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	tableBucketARN := fwflex.StringValueFromFramework(ctx, data.TableBucketARN)
	input := s3tables.DeleteTableBucketMetricsConfigurationInput{
		TableBucketARN: aws.String(tableBucketARN),
	}
	_, err := tfresource.RetryWhenIsA[*s3tables.DeleteTableBucketMetricsConfigurationOutput, *awstypes.ConflictException](ctx, tableBucketMetricsConfigurationPropagationTimeout, func(ctx context.Context) (*s3tables.DeleteTableBucketMetricsConfigurationOutput, error) {
		return conn.DeleteTableBucketMetricsConfiguration(ctx, &input)
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Tables Table Bucket Metrics Configuration (%s)", tableBucketARN), err.Error())

		return
	}
}

func findTableBucketMetricsConfigurationByARN(ctx context.Context, conn *s3tables.Client, tableBucketARN string) (*s3tables.GetTableBucketMetricsConfigurationOutput, error) {
	input := s3tables.GetTableBucketMetricsConfigurationInput{
		TableBucketARN: aws.String(tableBucketARN),
	}

	return findTableBucketMetricsConfiguration(ctx, conn, &input)
}

func findTableBucketMetricsConfiguration(ctx context.Context, conn *s3tables.Client, input *s3tables.GetTableBucketMetricsConfigurationInput) (*s3tables.GetTableBucketMetricsConfigurationOutput, error) {
	output, err := conn.GetTableBucketMetricsConfiguration(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || aws.ToString(output.TableBucketARN) == "" {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type tableBucketMetricsConfigurationResourceModel struct {
	framework.WithRegionModel
	MetricsConfigurationID types.String `tfsdk:"metrics_configuration_id"`
	TableBucketARN         fwtypes.ARN  `tfsdk:"table_bucket_arn"`
}
