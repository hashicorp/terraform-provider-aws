// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// See https://docs.aws.amazon.com/general/latest/gr/apprunner.html

var hostedZoneIDPerRegionMap = map[string]string{
	endpoints.UsEast2RegionID:      "Z0224347AD7KVHMLOX31",
	endpoints.UsEast1RegionID:      "Z01915732ZBZKC8D32TPT",
	endpoints.UsWest2RegionID:      "Z02243383FTQ64HJ5772Q",
	endpoints.ApSouth1RegionID:     "Z00855883LBHKTIC4ODF2",
	endpoints.ApSoutheast1RegionID: "Z09819469CZ3KQ8PWMCL",
	endpoints.ApSoutheast2RegionID: "Z03657752RA8799S0TI5I",
	endpoints.ApNortheast1RegionID: "Z08491812XW6IPYLR6CCA",
	endpoints.EuCentral1RegionID:   "Z0334911C2FDI2Q9M4FZ",
	endpoints.EuWest1RegionID:      "Z087551914Z2PCAU0QHMW",
	endpoints.EuWest2RegionID:      "Z098228427VC6B3IX76ON",
	endpoints.EuWest3RegionID:      "Z087117439MBKHYM69QS6",
}

// @FrameworkDataSource(name="Hosted Zone ID")
func newHostedZoneIDDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &hostedZoneIDDataSource{}, nil
}

type hostedZoneIDDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *hostedZoneIDDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_apprunner_hosted_zone_id"
}

func (d *hostedZoneIDDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrRegion: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

func (d *hostedZoneIDDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data hostedZoneIDDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	var region string
	if data.Region.IsNull() {
		region = d.Meta().Region
	} else {
		region = data.Region.ValueString()
	}

	if zoneID, ok := hostedZoneIDPerRegionMap[region]; ok {
		data.ID = types.StringValue(zoneID)
		data.Region = types.StringValue(region)
	} else {
		response.Diagnostics.AddError("unsupported AWS Region", fmt.Sprintf("region %s is currently not supported", region))

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type hostedZoneIDDataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	Region types.String `tfsdk:"region"`
}
