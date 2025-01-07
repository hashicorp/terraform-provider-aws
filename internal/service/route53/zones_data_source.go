// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_route53_zones", name="Zones")
func newZonesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &zonesDataSource{}, nil
}

type zonesDataSource struct {
	framework.DataSourceWithConfigure
}

func (*zonesDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_route53_zones"
}

func (d *zonesDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrIDs: schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *zonesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data zonesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().Route53Client(ctx)

	var zoneIDs []string
	input := &route53.ListHostedZonesInput{}
	pages := route53.NewListHostedZonesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			response.Diagnostics.AddError("listing Route 53 Hosted Zones", err.Error())

			return
		}

		for _, v := range page.HostedZones {
			zoneIDs = append(zoneIDs, cleanZoneID(aws.ToString(v.Id)))
		}
	}

	data.ID = types.StringValue(d.Meta().Region(ctx))
	data.ZoneIDs = fwflex.FlattenFrameworkStringValueList(ctx, zoneIDs)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type zonesDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	ZoneIDs types.List   `tfsdk:"ids"`
}
