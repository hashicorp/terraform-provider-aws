// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// See https://docs.aws.amazon.com/general/latest/gr/elb.html#elb_region.
var hostedZoneIDPerRegionMap = map[string]string{
	names.AFSouth1RegionID:     "Z268VQBMOI5EKX",
	names.APEast1RegionID:      "Z3DQVH9N71FHZ0",
	names.APNortheast1RegionID: "Z14GRHDCWA56QT",
	names.APNortheast2RegionID: "ZWKZPGTI48KDX",
	names.APNortheast3RegionID: "Z5LXEXXYW11ES",
	names.APSouth1RegionID:     "ZP97RAFLXTNZK",
	names.APSouth2RegionID:     "Z0173938T07WNTVAEPZN",
	names.APSoutheast1RegionID: "Z1LMS91P8CMLE5",
	names.APSoutheast2RegionID: "Z1GM3OXH4ZPM65",
	names.APSoutheast3RegionID: "Z08888821HLRG5A9ZRTER",
	names.APSoutheast4RegionID: "Z09517862IB2WZLPXG76F",
	names.CACentral1RegionID:   "ZQSVJUPU6J1EY",
	names.CAWest1RegionID:      "Z06473681N0SF6OS049SD",
	names.CNNorth1RegionID:     "Z1GDH35T77C1KE",
	names.CNNorthwest1RegionID: "ZM7IZAIOVVDZF",
	names.EUCentral1RegionID:   "Z215JYRZR1TBD5",
	names.EUCentral2RegionID:   "Z06391101F2ZOEP8P5EB3",
	names.EUNorth1RegionID:     "Z23TAZ6LKFMNIO",
	names.EUSouth1RegionID:     "Z3ULH7SSC9OV64",
	names.EUSouth2RegionID:     "Z0956581394HF5D5LXGAP",
	names.EUWest1RegionID:      "Z32O12XQLNTSW2",
	names.EUWest2RegionID:      "ZHURV8PSTC4K8",
	names.EUWest3RegionID:      "Z3Q77PNBQS71R4",
	names.ILCentral1RegionID:   "Z09170902867EHPV2DABU",
	names.MECentral1RegionID:   "Z08230872XQRWHG2XF6I",
	names.MESouth1RegionID:     "ZS929ML54UICD",
	names.SAEast1RegionID:      "Z2P70J7HTTTPLU",
	names.USEast1RegionID:      "Z35SXDOTRQ7X7K",
	names.USEast2RegionID:      "Z3AADJGX6KTTL2",
	names.USGovEast1RegionID:   "Z166TLBEWOO7G0",
	names.USGovWest1RegionID:   "Z33AYJ8TM3BH4J",
	names.USWest1RegionID:      "Z368ELLRRE2KJ0",
	names.USWest2RegionID:      "Z1H1FL5HABSF5",
}

// @SDKDataSource("aws_elb_hosted_zone_id", name="Hosted Zone ID")
func dataSourceHostedZoneID() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceHostedZoneIDRead,

		Schema: map[string]*schema.Schema{
			names.AttrRegion: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceHostedZoneIDRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	region := meta.(*conns.AWSClient).Region
	if v, ok := d.GetOk(names.AttrRegion); ok {
		region = v.(string)
	}

	if v, ok := hostedZoneIDPerRegionMap[region]; ok {
		d.SetId(v)
		return diags
	}

	return sdkdiag.AppendErrorf(diags, "unsupported AWS Region: %s", region)
}
