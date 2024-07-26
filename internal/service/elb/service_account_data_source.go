// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// See http://docs.aws.amazon.com/elasticloadbalancing/latest/classic/enable-access-logs.html#attach-bucket-policy
// See https://docs.aws.amazon.com/elasticloadbalancing/latest/application/load-balancer-access-logs.html#access-logging-bucket-permissions
var accountIDPerRegionMap = map[string]string{
	names.AFSouth1RegionID:     "098369216593",
	names.APEast1RegionID:      "754344448648",
	names.APNortheast1RegionID: "582318560864",
	names.APNortheast2RegionID: "600734575887",
	names.APNortheast3RegionID: "383597477331",
	names.APSouth1RegionID:     "718504428378",
	names.APSoutheast1RegionID: "114774131450",
	names.APSoutheast2RegionID: "783225319266",
	names.APSoutheast3RegionID: "589379963580",
	names.CACentral1RegionID:   "985666609251",
	names.CNNorth1RegionID:     "638102146993",
	names.CNNorthwest1RegionID: "037604701340",
	names.EUCentral1RegionID:   "054676820928",
	names.EUNorth1RegionID:     "897822967062",
	names.EUSouth1RegionID:     "635631232127",
	names.EUWest1RegionID:      "156460612806",
	names.EUWest2RegionID:      "652711504416",
	names.EUWest3RegionID:      "009996457667",
	// names.MECentral1RegionID:   "",
	names.MESouth1RegionID:   "076674570225",
	names.SAEast1RegionID:    "507241528517",
	names.USEast1RegionID:    "127311923021",
	names.USEast2RegionID:    "033677994240",
	names.USGovEast1RegionID: "190560391635",
	names.USGovWest1RegionID: "048591011584",
	names.USWest1RegionID:    "027434742980",
	names.USWest2RegionID:    "797873946194",
}

// @SDKDataSource("aws_elb_service_account", name="Service Account")
func dataSourceServiceAccount() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceServiceAccountRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrRegion: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceServiceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	region := meta.(*conns.AWSClient).Region
	if v, ok := d.GetOk(names.AttrRegion); ok {
		region = v.(string)
	}

	if v, ok := accountIDPerRegionMap[region]; ok {
		d.SetId(v)
		arn := arn.ARN{
			Partition: meta.(*conns.AWSClient).Partition,
			Service:   "iam",
			AccountID: v,
			Resource:  "root",
		}.String()
		d.Set(names.AttrARN, arn)

		return diags
	}

	return sdkdiag.AppendErrorf(diags, "unsupported AWS Region: %s", region)
}
