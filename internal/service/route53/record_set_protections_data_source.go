// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkDataSource("aws_route53_record_set_protections", name="Record Set Protections")
func newRecordSetProtectionsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &recordSetProtectionsDataSource{}, nil
}

type recordSetProtectionsDataSource struct {
	framework.DataSourceWithModel[recordSetProtectionsDataSourceModel]
}

func (d *recordSetProtectionsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"protections": framework.DataSourceComputedListOfObjectAttribute[recordSetProtectionModel](ctx),
			"zone_id": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *recordSetProtectionsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data recordSetProtectionsDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().Route53Client(ctx)

	zoneID := fwflex.StringValueFromFramework(ctx, data.ZoneID)
	output, err := findRecordSetProtectionsByZoneID(ctx, conn, zoneID)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, zoneID)
		return
	}

	// Return an empty list, not null, for a zone with no protections.
	if output == nil {
		output = []awstypes.ResourceRecordSetProtection{}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data.Protections))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

type recordSetProtectionsDataSourceModel struct {
	Protections fwtypes.ListNestedObjectValueOf[recordSetProtectionModel] `tfsdk:"protections"`
	ZoneID      types.String                                              `tfsdk:"zone_id"`
}

type recordSetProtectionModel struct {
	Comment    types.String                        `tfsdk:"comment"`
	RecordName types.String                        `tfsdk:"record_name"`
	RecordType fwtypes.StringEnum[awstypes.RRType] `tfsdk:"record_type"`
}
