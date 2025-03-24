// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb

import (
	"context"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// See https://docs.aws.amazon.com/general/latest/gr/elb.html#elb_region.
var hostedZoneIDPerRegionMap = map[string]string{
	endpoints.AfSouth1RegionID:     "Z268VQBMOI5EKX",
	endpoints.ApEast1RegionID:      "Z3DQVH9N71FHZ0",
	endpoints.ApNortheast1RegionID: "Z14GRHDCWA56QT",
	endpoints.ApNortheast2RegionID: "ZWKZPGTI48KDX",
	endpoints.ApNortheast3RegionID: "Z5LXEXXYW11ES",
	endpoints.ApSouth1RegionID:     "ZP97RAFLXTNZK",
	endpoints.ApSouth2RegionID:     "Z0173938T07WNTVAEPZN",
	endpoints.ApSoutheast1RegionID: "Z1LMS91P8CMLE5",
	endpoints.ApSoutheast2RegionID: "Z1GM3OXH4ZPM65",
	endpoints.ApSoutheast3RegionID: "Z08888821HLRG5A9ZRTER",
	endpoints.ApSoutheast4RegionID: "Z09517862IB2WZLPXG76F",
	endpoints.ApSoutheast5RegionID: "Z06010284QMVVW7WO5J",
	endpoints.ApSoutheast7RegionID: "Z0390008CMBRTHFGWBCB",
	endpoints.CaCentral1RegionID:   "ZQSVJUPU6J1EY",
	endpoints.CaWest1RegionID:      "Z06473681N0SF6OS049SD",
	endpoints.CnNorth1RegionID:     "Z1GDH35T77C1KE",
	endpoints.CnNorthwest1RegionID: "ZM7IZAIOVVDZF",
	endpoints.EuCentral1RegionID:   "Z215JYRZR1TBD5",
	endpoints.EuCentral2RegionID:   "Z06391101F2ZOEP8P5EB3",
	endpoints.EuNorth1RegionID:     "Z23TAZ6LKFMNIO",
	endpoints.EuSouth1RegionID:     "Z3ULH7SSC9OV64",
	endpoints.EuSouth2RegionID:     "Z0956581394HF5D5LXGAP",
	endpoints.EuWest1RegionID:      "Z32O12XQLNTSW2",
	endpoints.EuWest2RegionID:      "ZHURV8PSTC4K8",
	endpoints.EuWest3RegionID:      "Z3Q77PNBQS71R4",
	endpoints.IlCentral1RegionID:   "Z09170902867EHPV2DABU",
	endpoints.MeCentral1RegionID:   "Z08230872XQRWHG2XF6I",
	endpoints.MeSouth1RegionID:     "ZS929ML54UICD",
	endpoints.MxCentral1RegionID:   "Z023552324OKD1BB28BH5",
	endpoints.SaEast1RegionID:      "Z2P70J7HTTTPLU",
	endpoints.UsEast1RegionID:      "Z35SXDOTRQ7X7K",
	endpoints.UsEast2RegionID:      "Z3AADJGX6KTTL2",
	endpoints.UsGovEast1RegionID:   "Z166TLBEWOO7G0",
	endpoints.UsGovWest1RegionID:   "Z33AYJ8TM3BH4J",
	endpoints.UsWest1RegionID:      "Z368ELLRRE2KJ0",
	endpoints.UsWest2RegionID:      "Z1H1FL5HABSF5",
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

func dataSourceHostedZoneIDRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	region := meta.(*conns.AWSClient).Region(ctx)
	if v, ok := d.GetOk(names.AttrRegion); ok {
		region = v.(string)
	}

	if v, ok := hostedZoneIDPerRegionMap[region]; ok {
		d.SetId(v)
		return diags
	}

	return sdkdiag.AppendErrorf(diags, "unsupported AWS Region: %s", region)
}
