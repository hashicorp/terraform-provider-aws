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
)

// See https://docs.aws.amazon.com/general/latest/gr/apprunner.html

var HostedZoneIdPerRegionMap = map[string]string{
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

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource(name="Hosted Zone ID")
func newDataSourceHostedZoneID(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceHostedZoneID{}, nil
}

const (
	DSNameHostedZoneID = "Hosted Zone ID Data Source"
)

type dataSourceHostedZoneID struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceHostedZoneID) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_apprunner_hosted_zone_id"
}

func (d *dataSourceHostedZoneID) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"region": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

func (d *dataSourceHostedZoneID) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceHostedZoneIDData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var region string
	if data.Region.IsNull() {
		region = d.Meta().Region
	} else {
		region = data.Region.ValueString()
	}

	if zoneId, ok := HostedZoneIdPerRegionMap[region]; ok {
		data.ID = types.StringValue(zoneId)
		data.Region = types.StringValue(region)
	} else {
		resp.Diagnostics.AddError("unsupported AWS Region", fmt.Sprintf("region %s is currently not supported", region))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceHostedZoneIDData struct {
	ID     types.String `tfsdk:"id"`
	Region types.String `tfsdk:"region"`
}
