// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// See https://docs.aws.amazon.com/general/latest/gr/elb.html#elb_region
var hostedZoneIDPerRegionALBMap = map[string]string{
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

// See https://docs.aws.amazon.com/general/latest/gr/elb.html#elb_region
var hostedZoneIDPerRegionNLBMap = map[string]string{
	names.AFSouth1RegionID:     "Z203XCE67M25HM",
	names.APEast1RegionID:      "Z12Y7K3UBGUAD1",
	names.APNortheast1RegionID: "Z31USIVHYNEOWT",
	names.APNortheast2RegionID: "ZIBE1TIR4HY56",
	names.APNortheast3RegionID: "Z1GWIQ4HH19I5X",
	names.APSouth1RegionID:     "ZVDDRBQ08TROA",
	names.APSouth2RegionID:     "Z0711778386UTO08407HT",
	names.APSoutheast1RegionID: "ZKVM4W9LS7TM",
	names.APSoutheast2RegionID: "ZCT6FZBF4DROD",
	names.APSoutheast3RegionID: "Z01971771FYVNCOVWJU1G",
	names.APSoutheast4RegionID: "Z01156963G8MIIL7X90IV",
	names.CACentral1RegionID:   "Z2EPGBW3API2WT",
	names.CAWest1RegionID:      "Z02754302KBB00W2LKWZ9",
	names.CNNorth1RegionID:     "Z3QFB96KMJ7ED6",
	names.CNNorthwest1RegionID: "ZQEIKTCZ8352D",
	names.EUCentral1RegionID:   "Z3F0SRJ5LGBH90",
	names.EUCentral2RegionID:   "Z02239872DOALSIDCX66S",
	names.EUNorth1RegionID:     "Z1UDT6IFJ4EJM",
	names.EUSouth1RegionID:     "Z23146JA1KNAFP",
	names.EUSouth2RegionID:     "Z1011216NVTVYADP1SSV",
	names.EUWest1RegionID:      "Z2IFOLAFXWLO4F",
	names.EUWest2RegionID:      "ZD4D7Y8KGAS4G",
	names.EUWest3RegionID:      "Z1CMS0P5QUZ6D5",
	names.ILCentral1RegionID:   "Z0313266YDI6ZRHTGQY4",
	names.MECentral1RegionID:   "Z00282643NTTLPANJJG2P",
	names.MESouth1RegionID:     "Z3QSRYVP46NYYV",
	names.SAEast1RegionID:      "ZTK26PT1VY4CU",
	names.USEast1RegionID:      "Z26RNL4JYFTOTI",
	names.USEast2RegionID:      "ZLMOA37VPKANP",
	names.USGovEast1RegionID:   "Z1ZSMQQ6Q24QQ8",
	names.USGovWest1RegionID:   "ZMG1MZ2THAWF1",
	names.USWest1RegionID:      "Z24FKFUX50B4VW",
	names.USWest2RegionID:      "Z18D5FSROUN65G",
}

// @SDKDataSource("aws_lb_hosted_zone_id", name="Hosted Zone ID")
func dataSourceHostedZoneID() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceHostedZoneIDRead,

		Schema: map[string]*schema.Schema{
			names.AttrRegion: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidRegionName,
			},
			"load_balancer_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      awstypes.LoadBalancerTypeEnumApplication,
				ValidateFunc: validation.StringInSlice(enum.Slice[awstypes.LoadBalancerTypeEnum](awstypes.LoadBalancerTypeEnumApplication, awstypes.LoadBalancerTypeEnumNetwork), false),
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

	lbType := awstypes.LoadBalancerTypeEnumApplication
	if v, ok := d.GetOk("load_balancer_type"); ok {
		lbType = awstypes.LoadBalancerTypeEnum(v.(string))
	}

	switch lbType {
	case awstypes.LoadBalancerTypeEnumApplication:
		if v, ok := hostedZoneIDPerRegionALBMap[region]; ok {
			d.SetId(v)
		} else {
			return sdkdiag.AppendErrorf(diags, "unsupported AWS Region: %s", region)
		}
	case awstypes.LoadBalancerTypeEnumNetwork:
		if v, ok := hostedZoneIDPerRegionNLBMap[region]; ok {
			d.SetId(v)
		} else {
			return sdkdiag.AppendErrorf(diags, "unsupported AWS Region: %s", region)
		}
	}

	return diags
}
