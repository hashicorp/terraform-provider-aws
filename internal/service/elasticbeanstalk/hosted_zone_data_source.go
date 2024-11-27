// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk

import (
	"context"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// See https://docs.aws.amazon.com/general/latest/gr/elasticbeanstalk.html

var hostedZoneIDs = map[string]string{
	endpoints.AfSouth1RegionID: "Z1EI3BVKMKK4AM",
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
	names.EUNorth1RegionID:     "Z23GO28BZ5AETM",
	names.EUSouth1RegionID:     "Z10VDYYOA2JFKM",
	names.EUWest1RegionID:      "Z2NYPWQ7DFZAZH",
	names.EUWest2RegionID:      "Z1GKAAAUGATPF1",
	names.EUWest3RegionID:      "Z5WN6GAYWG5OB",
	names.ILCentral1RegionID:   "Z02941091PERNCB1MI5H7",
	// names.MECentral1RegionID:   "",
	names.MESouth1RegionID:   "Z2BBTEKR2I36N2",
	names.SAEast1RegionID:    "Z10X7K2B4QSOFV",
	names.USEast1RegionID:    "Z117KPS5GTRQ2G",
	names.USEast2RegionID:    "Z14LCN19Q5QHIC",
	names.USWest1RegionID:    "Z1LQECGX5PH1X",
	names.USWest2RegionID:    "Z38NKT9BP95V3O",
	names.USGovEast1RegionID: "Z35TSARG0EJ4VU",
	names.USGovWest1RegionID: "Z4KAURWC4UUUG",
}

// @SDKDataSource("aws_elastic_beanstalk_hosted_zone")
func dataSourceHostedZone() *schema.Resource {
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
	region := meta.(*conns.AWSClient).Region(ctx)
	if v, ok := d.GetOk(names.AttrRegion); ok {
		region = v.(string)
	}

	zoneID, ok := hostedZoneIDs[region]
	if !ok {
		return sdkdiag.AppendErrorf(diags, "unsupported Elastic Beanstalk Region (%s)", region)
	}

	d.SetId(zoneID)
	d.Set(names.AttrRegion, region)

	return diags
}
