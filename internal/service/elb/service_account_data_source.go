// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// See http://docs.aws.amazon.com/elasticloadbalancing/latest/classic/enable-access-logs.html#attach-bucket-policy
// See https://docs.aws.amazon.com/elasticloadbalancing/latest/application/load-balancer-access-logs.html#access-logging-bucket-permissions

var AccountIdPerRegionMap = map[string]string{
	endpoints.AfSouth1RegionID:     "098369216593",
	endpoints.ApEast1RegionID:      "754344448648",
	endpoints.ApNortheast1RegionID: "582318560864",
	endpoints.ApNortheast2RegionID: "600734575887",
	endpoints.ApNortheast3RegionID: "383597477331",
	endpoints.ApSouth1RegionID:     "718504428378",
	endpoints.ApSoutheast1RegionID: "114774131450",
	endpoints.ApSoutheast2RegionID: "783225319266",
	endpoints.ApSoutheast3RegionID: "589379963580",
	endpoints.CaCentral1RegionID:   "985666609251",
	endpoints.CnNorth1RegionID:     "638102146993",
	endpoints.CnNorthwest1RegionID: "037604701340",
	endpoints.EuCentral1RegionID:   "054676820928",
	endpoints.EuNorth1RegionID:     "897822967062",
	endpoints.EuSouth1RegionID:     "635631232127",
	endpoints.EuWest1RegionID:      "156460612806",
	endpoints.EuWest2RegionID:      "652711504416",
	endpoints.EuWest3RegionID:      "009996457667",
	// endpoints.MeCentral1RegionID:   "",
	endpoints.MeSouth1RegionID:   "076674570225",
	endpoints.SaEast1RegionID:    "507241528517",
	endpoints.UsEast1RegionID:    "127311923021",
	endpoints.UsEast2RegionID:    "033677994240",
	endpoints.UsGovEast1RegionID: "190560391635",
	endpoints.UsGovWest1RegionID: "048591011584",
	endpoints.UsWest1RegionID:    "027434742980",
	endpoints.UsWest2RegionID:    "797873946194",
}

// @SDKDataSource("aws_elb_service_account")
func DataSourceServiceAccount() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceServiceAccountRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceServiceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	region := meta.(*conns.AWSClient).Region
	if v, ok := d.GetOk("region"); ok {
		region = v.(string)
	}

	if accid, ok := AccountIdPerRegionMap[region]; ok {
		d.SetId(accid)
		arn := arn.ARN{
			Partition: meta.(*conns.AWSClient).Partition,
			Service:   "iam",
			AccountID: accid,
			Resource:  "root",
		}.String()
		d.Set("arn", arn)

		return diags
	}

	return sdkdiag.AppendErrorf(diags, "Unknown region (%q)", region)
}
