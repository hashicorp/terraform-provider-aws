package elbv2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// See https://docs.aws.amazon.com/general/latest/gr/elb.html#elb_region
var HostedZoneIdPerRegionALBMap = map[string]string{
	endpoints.AfSouth1RegionID:     "Z268VQBMOI5EKX",
	endpoints.ApEast1RegionID:      "Z3DQVH9N71FHZ0",
	endpoints.ApNortheast1RegionID: "Z14GRHDCWA56QT",
	endpoints.ApNortheast2RegionID: "ZWKZPGTI48KDX",
	endpoints.ApNortheast3RegionID: "Z5LXEXXYW11ES",
	endpoints.ApSouth1RegionID:     "ZP97RAFLXTNZK",
	endpoints.ApSoutheast1RegionID: "Z1LMS91P8CMLE5",
	endpoints.ApSoutheast2RegionID: "Z1GM3OXH4ZPM65",
	endpoints.ApSoutheast3RegionID: "Z08888821HLRG5A9ZRTER",
	endpoints.CaCentral1RegionID:   "ZQSVJUPU6J1EY",
	endpoints.CnNorth1RegionID:     "Z1GDH35T77C1KE",
	endpoints.CnNorthwest1RegionID: "ZM7IZAIOVVDZF",
	endpoints.EuCentral1RegionID:   "Z215JYRZR1TBD5",
	endpoints.EuNorth1RegionID:     "Z23TAZ6LKFMNIO",
	endpoints.EuSouth1RegionID:     "Z3ULH7SSC9OV64",
	endpoints.EuWest1RegionID:      "Z32O12XQLNTSW2",
	endpoints.EuWest2RegionID:      "ZHURV8PSTC4K8",
	endpoints.EuWest3RegionID:      "Z3Q77PNBQS71R4",
	endpoints.MeSouth1RegionID:     "ZS929ML54UICD",
	endpoints.SaEast1RegionID:      "Z2P70J7HTTTPLU",
	endpoints.UsEast1RegionID:      "Z35SXDOTRQ7X7K",
	endpoints.UsEast2RegionID:      "Z3AADJGX6KTTL2",
	endpoints.UsGovEast1RegionID:   "Z166TLBEWOO7G0",
	endpoints.UsGovWest1RegionID:   "Z33AYJ8TM3BH4J",
	endpoints.UsWest1RegionID:      "Z368ELLRRE2KJ0",
	endpoints.UsWest2RegionID:      "Z1H1FL5HABSF5",
}

// See https://docs.aws.amazon.com/general/latest/gr/elb.html#elb_region
var HostedZoneIdPerRegionNLBMap = map[string]string{
	endpoints.AfSouth1RegionID:     "Z203XCE67M25HM",
	endpoints.ApEast1RegionID:      "Z12Y7K3UBGUAD1",
	endpoints.ApNortheast1RegionID: "Z31USIVHYNEOWT",
	endpoints.ApNortheast2RegionID: "ZIBE1TIR4HY56",
	endpoints.ApNortheast3RegionID: "Z1GWIQ4HH19I5X",
	endpoints.ApSouth1RegionID:     "ZVDDRBQ08TROA",
	endpoints.ApSoutheast1RegionID: "ZKVM4W9LS7TM",
	endpoints.ApSoutheast2RegionID: "ZCT6FZBF4DROD",
	endpoints.ApSoutheast3RegionID: "Z01971771FYVNCOVWJU1G",
	endpoints.CaCentral1RegionID:   "Z2EPGBW3API2WT",
	endpoints.CnNorth1RegionID:     "Z3QFB96KMJ7ED6",
	endpoints.CnNorthwest1RegionID: "ZQEIKTCZ8352D",
	endpoints.EuCentral1RegionID:   "Z3F0SRJ5LGBH90",
	endpoints.EuNorth1RegionID:     "Z1UDT6IFJ4EJM",
	endpoints.EuSouth1RegionID:     "Z23146JA1KNAFP",
	endpoints.EuWest1RegionID:      "Z2IFOLAFXWLO4F",
	endpoints.EuWest2RegionID:      "ZD4D7Y8KGAS4G",
	endpoints.EuWest3RegionID:      "Z1CMS0P5QUZ6D5",
	endpoints.MeSouth1RegionID:     "Z3QSRYVP46NYYV",
	endpoints.SaEast1RegionID:      "ZTK26PT1VY4CU",
	endpoints.UsEast1RegionID:      "Z26RNL4JYFTOTI",
	endpoints.UsEast2RegionID:      "ZLMOA37VPKANP",
	endpoints.UsGovEast1RegionID:   "Z1ZSMQQ6Q24QQ8",
	endpoints.UsGovWest1RegionID:   "ZMG1MZ2THAWF1",
	endpoints.UsWest1RegionID:      "Z24FKFUX50B4VW",
	endpoints.UsWest2RegionID:      "Z18D5FSROUN65G",
}

func DataSourceHostedZoneID() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHostedZoneIDRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidRegionName,
			},
			"load_balancer_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      elbv2.LoadBalancerTypeEnumApplication,
				ValidateFunc: validation.StringInSlice([]string{elbv2.LoadBalancerTypeEnumApplication, elbv2.LoadBalancerTypeEnumNetwork}, false),
			},
		},
	}
}

func dataSourceHostedZoneIDRead(d *schema.ResourceData, meta interface{}) error {
	region := meta.(*conns.AWSClient).Region
	if v, ok := d.GetOk("region"); ok {
		region = v.(string)
	}

	lbType := elbv2.LoadBalancerTypeEnumApplication
	if v, ok := d.GetOk("load_balancer_type"); ok {
		lbType = v.(string)
	}

	if lbType == elbv2.LoadBalancerTypeEnumApplication {
		if zoneId, ok := HostedZoneIdPerRegionALBMap[region]; ok {
			d.SetId(zoneId)
		} else {
			return fmt.Errorf("unsupported AWS Region: %s", region)
		}
	} else if lbType == elbv2.LoadBalancerTypeEnumNetwork {
		if zoneId, ok := HostedZoneIdPerRegionNLBMap[region]; ok {
			d.SetId(zoneId)
		} else {
			return fmt.Errorf("unsupported AWS Region: %s", region)
		}
	}

	return nil
}
