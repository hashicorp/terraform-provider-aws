// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// See https://docs.aws.amazon.com/general/latest/gr/elasticbeanstalk.html

var HostedZoneIDs = map[string]string{
	endpoints.AfSouth1RegionID:     "Z1EI3BVKMKK4AM",
	endpoints.ApSoutheast1RegionID: "Z16FZ9L249IFLT",
	endpoints.ApSoutheast2RegionID: "Z2PCDNR3VC2G1N",
	endpoints.ApSoutheast3RegionID: "Z05913172VM7EAZB40TA8",
	endpoints.ApEast1RegionID:      "ZPWYUBWRU171A",
	endpoints.ApNortheast1RegionID: "Z1R25G3KIG2GBW",
	endpoints.ApNortheast2RegionID: "Z3JE5OI70TWKCP",
	endpoints.ApNortheast3RegionID: "ZNE5GEY1TIAGY",
	endpoints.ApSouth1RegionID:     "Z18NTBI3Y7N9TZ",
	endpoints.CaCentral1RegionID:   "ZJFCZL7SSZB5I",
	endpoints.EuCentral1RegionID:   "Z1FRNW7UH4DEZJ",
	endpoints.EuNorth1RegionID:     "Z23GO28BZ5AETM",
	endpoints.EuSouth1RegionID:     "Z10VDYYOA2JFKM",
	endpoints.EuWest1RegionID:      "Z2NYPWQ7DFZAZH",
	endpoints.EuWest2RegionID:      "Z1GKAAAUGATPF1",
	endpoints.EuWest3RegionID:      "Z5WN6GAYWG5OB",
	endpoints.IlCentral1RegionID:   "Z02941091PERNCB1MI5H7",
	// endpoints.MeCentral1RegionID:   "",
	endpoints.MeSouth1RegionID:   "Z2BBTEKR2I36N2",
	endpoints.SaEast1RegionID:    "Z10X7K2B4QSOFV",
	endpoints.UsEast1RegionID:    "Z117KPS5GTRQ2G",
	endpoints.UsEast2RegionID:    "Z14LCN19Q5QHIC",
	endpoints.UsWest1RegionID:    "Z1LQECGX5PH1X",
	endpoints.UsWest2RegionID:    "Z38NKT9BP95V3O",
	endpoints.UsGovEast1RegionID: "Z35TSARG0EJ4VU",
	endpoints.UsGovWest1RegionID: "Z4KAURWC4UUUG",
}

// @SDKDataSource("aws_elastic_beanstalk_hosted_zone")
func DataSourceHostedZone() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceHostedZoneRead,

		Schema: map[string]*schema.Schema{
			names.AttrRegion: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceHostedZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	region := meta.(*conns.AWSClient).Region
	if v, ok := d.GetOk(names.AttrRegion); ok {
		region = v.(string)
	}

	zoneID, ok := HostedZoneIDs[region]

	if !ok {
		return sdkdiag.AppendErrorf(diags, "Unsupported region: %s", region)
	}

	d.SetId(zoneID)
	d.Set(names.AttrRegion, region)
	return diags
}
