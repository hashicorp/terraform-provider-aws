// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// See http://docs.aws.amazon.com/redshift/latest/mgmt/db-auditing.html#db-auditing-bucket-permissions
// See https://docs.aws.amazon.com/govcloud-us/latest/UserGuide/govcloud-redshift.html
// See https://docs.amazonaws.cn/en_us/redshift/latest/mgmt/db-auditing.html#db-auditing-bucket-permissions

var ServiceAccountPerRegionMap = map[string]string{
	names.AFSouth1RegionID:     "365689465814",
	names.APEast1RegionID:      "313564881002",
	names.APNortheast1RegionID: "404641285394",
	names.APNortheast2RegionID: "760740231472",
	names.APNortheast3RegionID: "090321488786",
	names.APSouth1RegionID:     "865932855811",
	names.APSoutheast1RegionID: "361669875840",
	names.APSoutheast2RegionID: "762762565011",
	names.CACentral1RegionID:   "907379612154",
	names.CNNorth1RegionID:     "111890595117",
	names.CNNorthwest1RegionID: "660998842044",
	names.EUCentral1RegionID:   "053454850223",
	names.EUNorth1RegionID:     "729911121831",
	names.EUSouth1RegionID:     "945612479654",
	names.EUWest1RegionID:      "210876761215",
	names.EUWest2RegionID:      "307160386991",
	names.EUWest3RegionID:      "915173422425",
	// names.MECentral1RegionID:   "",
	names.MESouth1RegionID:   "013126148197",
	names.SAEast1RegionID:    "075028567923",
	names.USEast1RegionID:    "193672423079",
	names.USEast2RegionID:    "391106570357",
	names.USGovEast1RegionID: "665727464434",
	names.USGovWest1RegionID: "665727464434",
	names.USWest1RegionID:    "262260360010",
	names.USWest2RegionID:    "902366379725",
}

// @SDKDataSource("aws_redshift_service_account", name="Service Account")
func dataSourceServiceAccount() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceServiceAccountRead,

		Schema: map[string]*schema.Schema{
			names.AttrRegion: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		DeprecationMessage: `The aws_redshift_service_account data source has been deprecated and will be removed in a future version. Use a service principal name instead of AWS account ID in any relevant IAM policy.`,
	}
}

func dataSourceServiceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	region := meta.(*conns.AWSClient).Region
	if v, ok := d.GetOk(names.AttrRegion); ok {
		region = v.(string)
	}

	if accid, ok := ServiceAccountPerRegionMap[region]; ok {
		d.SetId(accid)
		arn := arn.ARN{
			Partition: meta.(*conns.AWSClient).Partition,
			Service:   "iam",
			AccountID: accid,
			Resource:  "user/logs",
		}.String()
		d.Set(names.AttrARN, arn)

		return diags
	}

	return sdkdiag.AppendErrorf(diags, "Unknown region (%q)", region)
}
