// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudtrail

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// See http://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-supported-regions.html
// See https://docs.aws.amazon.com/govcloud-us/latest/ug-east/verifying-cloudtrail.html
// See https://docs.aws.amazon.com/govcloud-us/latest/ug-west/verifying-cloudtrail.html

var serviceAccountPerRegionMap = map[string]string{
	names.AFSouth1RegionID:     "525921808201",
	names.APEast1RegionID:      "119688915426",
	names.APNortheast1RegionID: "216624486486",
	names.APNortheast2RegionID: "492519147666",
	names.APNortheast3RegionID: "765225791966",
	names.APSouth1RegionID:     "977081816279",
	names.APSouth2RegionID:     "582488909970",
	names.APSoutheast1RegionID: "903692715234",
	names.APSoutheast2RegionID: "284668455005",
	names.APSoutheast3RegionID: "069019280451",
	names.APSoutheast4RegionID: "187074758985",
	names.CACentral1RegionID:   "819402241893",
	names.CNNorth1RegionID:     "193415116832",
	names.CNNorthwest1RegionID: "681348832753",
	names.EUCentral1RegionID:   "035351147821",
	names.EUCentral2RegionID:   "453052556044",
	names.EUNorth1RegionID:     "829690693026",
	names.EUSouth1RegionID:     "669305197877",
	names.EUSouth2RegionID:     "757211635381",
	names.EUWest1RegionID:      "859597730677",
	names.EUWest2RegionID:      "282025262664",
	names.EUWest3RegionID:      "262312530599",
	names.ILCentral1RegionID:   "683224464357",
	names.MECentral1RegionID:   "585772288577",
	names.MESouth1RegionID:     "034638983726",
	names.SAEast1RegionID:      "814480443879",
	names.USEast1RegionID:      "086441151436",
	names.USEast2RegionID:      "475085895292",
	names.USGovEast1RegionID:   "608710470296",
	names.USGovWest1RegionID:   "608710470296",
	names.USWest1RegionID:      "388731089494",
	names.USWest2RegionID:      "113285607260",
}

// @SDKDataSource("aws_cloudtrail_service_account", name="Service Account")
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

	if v, ok := serviceAccountPerRegionMap[region]; ok {
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

	return sdkdiag.AppendErrorf(diags, "unsupported CloudTrail Service Account Region (%s)", region)
}
