package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

// See https://docs.aws.amazon.com/general/latest/gr/rande.html#elb_region
// See https://docs.amazonaws.cn/en_us/general/latest/gr/rande.html#elb_region
// First index is Application/Classic LB id's, second index is NLB id's
var elbHostedZoneIdPerRegionMap = map[string][]string{
	"ap-east-1":      {"Z3DQVH9N71FHZ0", "Z12Y7K3UBGUAD1"},
	"ap-northeast-1": {"Z14GRHDCWA56QT", "Z31USIVHYNEOWT"},
	"ap-northeast-2": {"ZWKZPGTI48KDX", "ZIBE1TIR4HY56"},
	"ap-northeast-3": {"Z5LXEXXYW11ES", "Z1GWIQ4HH19I5X"},
	"ap-south-1":     {"ZP97RAFLXTNZK", "ZVDDRBQ08TROA"},
	"ap-southeast-1": {"Z1LMS91P8CMLE5", "ZKVM4W9LS7TM"},
	"ap-southeast-2": {"Z1GM3OXH4ZPM65", "ZCT6FZBF4DROD"},
	"ca-central-1":   {"ZQSVJUPU6J1EY", "Z2EPGBW3API2WT"},
	"cn-north-1":     {"3BX2TMKNYI13Y", "Z3QFB96KMJ7ED6"},
	"cn-northwest-1": {"Z3BX2TMKNYI13Y", "ZQEIKTCZ8352D"},
	"eu-central-1":   {"Z215JYRZR1TBD5", "Z3F0SRJ5LGBH90"},
	"eu-north-1":     {"Z23TAZ6LKFMNIO", "Z1UDT6IFJ4EJM"},
	"eu-west-1":      {"Z32O12XQLNTSW2", "Z2IFOLAFXWLO4F"},
	"eu-west-2":      {"ZHURV8PSTC4K8", "ZD4D7Y8KGAS4G"},
	"eu-west-3":      {"Z3Q77PNBQS71R4", "Z1CMS0P5QUZ6D5"},
	"me-south-1":     {"ZS929ML54UICD", "Z3QSRYVP46NYYV"},
	"sa-east-1":      {"Z2P70J7HTTTPLU", "ZTK26PT1VY4CU"},
	"us-east-1":      {"Z35SXDOTRQ7X7K", "Z26RNL4JYFTOTI"},
	"us-east-2":      {"Z3AADJGX6KTTL2", "ZLMOA37VPKANP"},
	"us-west-1":      {"Z368ELLRRE2KJ0", "Z24FKFUX50B4VW"},
	"us-west-2":      {"Z1H1FL5HABSF5", "Z18D5FSROUN65G"},
}

var validElbTypes = []string{
	// elbv2 and elb don't have an option for classic
	elbv2.LoadBalancerTypeEnumApplication,
	"classic",
	elbv2.LoadBalancerTypeEnumNetwork,
}

func dataSourceAwsElbHostedZoneId() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsElbHostedZoneIdRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"elb_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(validElbTypes, false),
			},
		},
	}
}

func dataSourceAwsElbHostedZoneIdRead(d *schema.ResourceData, meta interface{}) error {
	region := meta.(*AWSClient).region
	if v, ok := d.GetOk("region"); ok {
		region = v.(string)
	}
	var i int
	if v, ok := d.GetOk("elb_type"); ok {
		if v.(string) == "classic" || v.(string) == "application" {
			i = 0
		} else if v.(string) == "network" {
			i = 1
		}
	} else {
		i = 0
	}
	if zoneId, ok := elbHostedZoneIdPerRegionMap[region]; ok {
		d.SetId(zoneId[i])
		return nil
	}

	return fmt.Errorf("Unknown region (%q)", region)
}
