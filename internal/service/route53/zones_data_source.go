// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_route53_zones", name="Zones")
func newZonesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &zonesDataSource{}, nil
}

type zonesDataSource struct {
	framework.DataSourceWithConfigure
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

	input := route53.ListHostedZonesInput{}
	output, err := findHostedZones(ctx, conn, &input, tfslices.PredicateTrue[*awstypes.HostedZone]())

	if err != nil {
		response.Diagnostics.AddError("reading Route 53 Hosted Zones", err.Error())

		return
	}

	zoneIDs := tfslices.ApplyToAll(output, func(v awstypes.HostedZone) string {
		return cleanZoneID(aws.ToString(v.Id))
	})

	data.ID = types.StringValue(d.Meta().Region(ctx))
	data.ZoneIDs = fwflex.FlattenFrameworkStringValueList(ctx, zoneIDs)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type zonesDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	ZoneIDs types.List   `tfsdk:"ids"`
}
