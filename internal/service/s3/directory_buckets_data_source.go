// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_s3_directory_buckets", name="Directory Buckets")
func newDirectoryBucketsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &directoryBucketsDataSource{}

	return d, nil
}

type directoryBucketsDataSource struct {
	framework.DataSourceWithModel[directoryBucketsDataSourceModel]
}

func (d *directoryBucketsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARNs: schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"buckets": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrID: framework.IDAttribute(),
		},
	}
}

func (d *directoryBucketsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data directoryBucketsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().S3ExpressClient(ctx)

	input := &s3.ListDirectoryBucketsInput{}
	var buckets []string
	pages := s3.NewListDirectoryBucketsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			response.Diagnostics.AddError("listing S3 Directory Buckets", err.Error())

			return
		}

		for _, v := range page.Buckets {
			buckets = append(buckets, aws.ToString(v.Name))
		}
	}

	data.ARNs = fwflex.FlattenFrameworkStringValueListOfString(ctx, tfslices.ApplyToAll(buckets, func(v string) string {
		return d.Meta().RegionalARN(ctx, "s3express", "bucket/"+v)
	}))
	data.Buckets = fwflex.FlattenFrameworkStringValueListOfString(ctx, buckets)
	data.ID = types.StringValue(d.Meta().Region(ctx))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type directoryBucketsDataSourceModel struct {
	framework.WithRegionModel
	ARNs    fwtypes.ListOfString `tfsdk:"arns"`
	Buckets fwtypes.ListOfString `tfsdk:"buckets"`
	ID      types.String         `tfsdk:"id"`
}
