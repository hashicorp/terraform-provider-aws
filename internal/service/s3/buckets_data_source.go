// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_s3_buckets", name="Buckets")
func newBucketsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &bucketsDataSource{}, nil
}

type bucketsDataSource struct {
	framework.DataSourceWithModel[bucketsDataSourceModel]
}

func (d *bucketsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"buckets": framework.DataSourceComputedListOfObjectAttribute[bucketsModel](ctx),
			"max_buckets": schema.Int32Attribute{
				Optional: true,
			},
			names.AttrPrefix: schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (d *bucketsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().S3Client(ctx)

	var data bucketsDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var input s3.ListBucketsInput
	if !data.Region.IsNull() {
		input.BucketRegion = data.Region.ValueStringPointer()
	}
	if !data.Prefix.IsNull() {
		input.Prefix = data.Prefix.ValueStringPointer()
	}
	if !data.MaxBuckets.IsNull() {
		input.MaxBuckets = data.MaxBuckets.ValueInt32Pointer()
	}

	out, err := findBuckets(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &data.Buckets))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

// findBuckets collects all buckets matching the given input by iterating
// over the listBuckets iterator. The caller can extend this function with
// Terraform-side filtering before the return if needed in the future.
func findBuckets(ctx context.Context, conn *s3.Client, input *s3.ListBucketsInput) ([]awstypes.Bucket, error) {
	var output []awstypes.Bucket

	for item, err := range listBuckets(ctx, conn, input) {
		if err != nil {
			return nil, err
		}
		output = append(output, item)
	}

	return output, nil
}

type bucketsDataSourceModel struct {
	framework.WithRegionModel
	Buckets    fwtypes.ListNestedObjectValueOf[bucketsModel] `tfsdk:"buckets"`
	MaxBuckets types.Int32                                   `tfsdk:"max_buckets"`
	Prefix     types.String                                  `tfsdk:"prefix"`
}

type bucketsModel struct {
	BucketArn    types.String      `tfsdk:"bucket_arn"`
	BucketRegion types.String      `tfsdk:"bucket_region"`
	CreationDate timetypes.RFC3339 `tfsdk:"creation_date"`
	Name         types.String      `tfsdk:"name"`
}
