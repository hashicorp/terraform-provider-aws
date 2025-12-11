// Copyright IBM Corp. 2014, 2025
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
	endpoints.ApEast1RegionID:  "ZPWYUBWRU171A",
	// endpoints.ApEast2RegionID:      "",
	endpoints.ApNortheast1RegionID: "Z1R25G3KIG2GBW",
	endpoints.ApNortheast2RegionID: "Z3JE5OI70TWKCP",
	endpoints.ApNortheast3RegionID: "ZNE5GEY1TIAGY",
	endpoints.ApSouth1RegionID:     "Z18NTBI3Y7N9TZ",
	// endpoints.ApSouth2RegionID:     "",
	endpoints.ApSoutheast1RegionID: "Z16FZ9L249IFLT",
	endpoints.ApSoutheast2RegionID: "Z2PCDNR3VC2G1N",
	endpoints.ApSoutheast3RegionID: "Z05913172VM7EAZB40TA8",
	// endpoints.ApSoutheast4RegionID: "",
	endpoints.ApSoutheast5RegionID: "Z18NTBI3Y7N9TZ",
	// endpoints.ApSoutheast6RegionID: "",
	endpoints.ApSoutheast7RegionID: "Z1R25G3KIG2GBW",
	endpoints.CaCentral1RegionID:   "ZJFCZL7SSZB5I",
	// endpoints.CaWest1RegionID:      "",
	// endpoints.CnNorth1RegionID:     "",
	// endpoints.CnNorthwest1RegionID: "",
	endpoints.EuCentral1RegionID: "Z1FRNW7UH4DEZJ",
	// endpoints.EuCentral2RegionID:   "",
	endpoints.EuNorth1RegionID:   "Z23GO28BZ5AETM",
	endpoints.EuSouth1RegionID:   "Z10VDYYOA2JFKM",
	endpoints.EuSouth2RegionID:   "Z23GO28BZ5AETM",
	endpoints.EuWest1RegionID:    "Z2NYPWQ7DFZAZH",
	endpoints.EuWest2RegionID:    "Z1GKAAAUGATPF1",
	endpoints.EuWest3RegionID:    "Z5WN6GAYWG5OB",
	endpoints.IlCentral1RegionID: "Z02941091PERNCB1MI5H7",
	endpoints.MeCentral1RegionID: "Z10X7K2B4QSOFV",
	endpoints.MeSouth1RegionID:   "Z2BBTEKR2I36N2",
	// endpoints.MxCentral1RegionID: "",
	endpoints.SaEast1RegionID:    "Z10X7K2B4QSOFV",
	endpoints.UsEast1RegionID:    "Z117KPS5GTRQ2G",
	endpoints.UsEast2RegionID:    "Z14LCN19Q5QHIC",
	endpoints.UsGovEast1RegionID: "Z35TSARG0EJ4VU",
	endpoints.UsGovWest1RegionID: "Z4KAURWC4UUUG",
	endpoints.UsWest1RegionID:    "Z1LQECGX5PH1X",
	endpoints.UsWest2RegionID:    "Z38NKT9BP95V3O",
}

// @SDKDataSource("aws_elastic_beanstalk_hosted_zone", name="Hosted Zone")
// @Region(validateOverrideInPartition=false)
func dataSourceHostedZone() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceHostedZoneRead,

		Schema: map[string]*schema.Schema{},
	}
}

func dataSourceHostedZoneRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	region := meta.(*conns.AWSClient).Region(ctx)
	zoneID, ok := hostedZoneIDs[region]
	if !ok {
		return sdkdiag.AppendErrorf(diags, "unsupported Elastic Beanstalk Region (%s)", region)
	}

	d.SetId(zoneID)
	d.Set(names.AttrRegion, region)

	return diags
}
