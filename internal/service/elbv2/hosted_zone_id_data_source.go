// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
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

// See https://docs.aws.amazon.com/general/latest/gr/elb.html#elb_region
var hostedZoneIDPerRegionNLBMap = map[string]string{
	endpoints.AfSouth1RegionID:     "Z203XCE67M25HM",
	endpoints.ApEast1RegionID:      "Z12Y7K3UBGUAD1",
	endpoints.ApNortheast1RegionID: "Z31USIVHYNEOWT",
	endpoints.ApNortheast2RegionID: "ZIBE1TIR4HY56",
	endpoints.ApNortheast3RegionID: "Z1GWIQ4HH19I5X",
	endpoints.ApSouth1RegionID:     "ZVDDRBQ08TROA",
	endpoints.ApSouth2RegionID:     "Z0711778386UTO08407HT",
	endpoints.ApSoutheast1RegionID: "ZKVM4W9LS7TM",
	endpoints.ApSoutheast2RegionID: "ZCT6FZBF4DROD",
	endpoints.ApSoutheast3RegionID: "Z01971771FYVNCOVWJU1G",
	endpoints.ApSoutheast4RegionID: "Z01156963G8MIIL7X90IV",
	endpoints.ApSoutheast5RegionID: "Z026317210H9ACVTRO6FB",
	endpoints.ApSoutheast7RegionID: "Z054363131YWATEMWRG5L",
	endpoints.CaCentral1RegionID:   "Z2EPGBW3API2WT",
	endpoints.CaWest1RegionID:      "Z02754302KBB00W2LKWZ9",
	endpoints.CnNorth1RegionID:     "Z3QFB96KMJ7ED6",
	endpoints.CnNorthwest1RegionID: "ZQEIKTCZ8352D",
	endpoints.EuCentral1RegionID:   "Z3F0SRJ5LGBH90",
	endpoints.EuCentral2RegionID:   "Z02239872DOALSIDCX66S",
	endpoints.EuNorth1RegionID:     "Z1UDT6IFJ4EJM",
	endpoints.EuSouth1RegionID:     "Z23146JA1KNAFP",
	endpoints.EuSouth2RegionID:     "Z1011216NVTVYADP1SSV",
	endpoints.EuWest1RegionID:      "Z2IFOLAFXWLO4F",
	endpoints.EuWest2RegionID:      "ZD4D7Y8KGAS4G",
	endpoints.EuWest3RegionID:      "Z1CMS0P5QUZ6D5",
	endpoints.IlCentral1RegionID:   "Z0313266YDI6ZRHTGQY4",
	endpoints.MeCentral1RegionID:   "Z00282643NTTLPANJJG2P",
	endpoints.MeSouth1RegionID:     "Z3QSRYVP46NYYV",
	endpoints.MxCentral1RegionID:   "Z02031231H3ID6HYJ9A7U",
	endpoints.SaEast1RegionID:      "ZTK26PT1VY4CU",
	endpoints.UsEast1RegionID:      "Z26RNL4JYFTOTI",
	endpoints.UsEast2RegionID:      "ZLMOA37VPKANP",
	endpoints.UsGovEast1RegionID:   "Z1ZSMQQ6Q24QQ8",
	endpoints.UsGovWest1RegionID:   "ZMG1MZ2THAWF1",
	endpoints.UsWest1RegionID:      "Z24FKFUX50B4VW",
	endpoints.UsWest2RegionID:      "Z18D5FSROUN65G",
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

func dataSourceHostedZoneIDRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	region := meta.(*conns.AWSClient).Region(ctx)
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
