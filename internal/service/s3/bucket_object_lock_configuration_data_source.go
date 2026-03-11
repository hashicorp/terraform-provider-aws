// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package s3

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_s3_bucket_object_lock_configuration", name="Bucket Object Lock Configuration")
func newBucketObjectLockConfigurationDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &bucketObjectLockConfigurationDataSource{}, nil
}

const (
	DSNameBucketObjectLockConfiguration = "Bucket Object Lock Configuration Data Source"
)

type bucketObjectLockConfigurationDataSource struct {
	framework.DataSourceWithModel[bucketObjectLockConfigurationDataSourceModel]
}

func (d *bucketObjectLockConfigurationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrBucket: schema.StringAttribute{
				Required: true,
			},
			names.AttrExpectedBucketOwner: schema.StringAttribute{
				Optional: true,
			},
			"object_lock_enabled": schema.StringAttribute{
				Computed: true,
			},
			names.AttrRule: framework.DataSourceComputedListOfObjectAttribute[dataBucketObjectLockConfigRuleModel](ctx),
		},
	}
}

func (d *bucketObjectLockConfigurationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().S3Client(ctx)

	var data bucketObjectLockConfigurationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bucket := data.Bucket.ValueString()
	if isDirectoryBucket(bucket) {
		conn = d.Meta().S3ExpressClient(ctx)
	}

	expectedBucketOwner := data.ExpectedBucketOwner.ValueString()
	out, err := findObjectLockConfiguration(ctx, conn, bucket, expectedBucketOwner)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3, create.ErrActionReading, DSNameBucketObjectLockConfiguration, bucket, err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type bucketObjectLockConfigurationDataSourceModel struct {
	framework.WithRegionModel
	Bucket              types.String                                                         `tfsdk:"bucket"`
	ExpectedBucketOwner types.String                                                         `tfsdk:"expected_bucket_owner"`
	ObjectLockEnabled   types.String                                                         `tfsdk:"object_lock_enabled"`
	Rule                fwtypes.ListNestedObjectValueOf[dataBucketObjectLockConfigRuleModel] `tfsdk:"rule"`
}

type dataBucketObjectLockConfigRuleModel struct {
	DefaultRetention fwtypes.ListNestedObjectValueOf[dataBucketObjectLockConfigDefaultRetentionModel] `tfsdk:"default_retention"`
}

type dataBucketObjectLockConfigDefaultRetentionModel struct {
	Days  types.Int64  `tfsdk:"days"`
	Mode  types.String `tfsdk:"mode"`
	Years types.Int64  `tfsdk:"years"`
}
