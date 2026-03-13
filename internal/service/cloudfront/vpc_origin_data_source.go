// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_cloudfront_vpc_origin", name="VPC Origin")
func newDataSourceVPCOrigin(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &dataSourceVPCOrigin{}
	return d, nil
}

const (
	DSNameVPCOrigin = "VPC Origin Data Source"
)

type dataSourceVPCOrigin struct {
	framework.DataSourceWithModel[dataSourceVPCOriginModel]
}

func (d *dataSourceVPCOrigin) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"etag": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *dataSourceVPCOrigin) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data dataSourceVPCOriginModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().CloudFrontClient(ctx)

	output, err := findVPCOriginByID(ctx, conn, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionReading, DSNameVPCOrigin, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	data.ARN = fwflex.StringToFramework(ctx, output.VpcOrigin.Arn)
	data.ETag = fwflex.StringToFramework(ctx, output.ETag)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceVPCOriginModel struct {
	ARN  types.String `tfsdk:"arn"`
	ETag types.String `tfsdk:"etag"`
	ID   types.String `tfsdk:"id"`
}
