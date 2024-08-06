// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_s3control_multi_region_access_point")
func dataSourceMultiRegionAccessPoint() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceMultiRegionAccessPointBlockRead,

		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			names.AttrAlias: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"public_access_block": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"block_public_acls": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"block_public_policy": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"ignore_public_acls": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"restrict_public_buckets": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"regions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrBucket: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"bucket_account_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrRegion: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceMultiRegionAccessPointBlockRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk(names.AttrAccountID); ok {
		accountID = v.(string)
	}
	name := d.Get(names.AttrName).(string)

	accessPoint, err := findMultiRegionAccessPointByTwoPartKey(ctx, conn, accountID, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Multi Region Access Point (%s): %s", name, err)
	}

	d.SetId(MultiRegionAccessPointCreateResourceID(accountID, name))

	alias := aws.ToString(accessPoint.Alias)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "s3",
		AccountID: accountID,
		Resource:  fmt.Sprintf("accesspoint/%s", alias),
	}.String()
	d.Set(names.AttrAccountID, accountID)
	d.Set(names.AttrAlias, alias)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrCreatedAt, aws.ToTime(accessPoint.CreatedAt).Format(time.RFC3339))
	// https://docs.aws.amazon.com/AmazonS3/latest/userguide//MultiRegionAccessPointRequests.html#MultiRegionAccessPointHostnames.
	d.Set(names.AttrDomainName, meta.(*conns.AWSClient).PartitionHostname(ctx, alias+".accesspoint.s3-global"))
	d.Set(names.AttrName, accessPoint.Name)
	d.Set("public_access_block", []interface{}{flattenPublicAccessBlockConfiguration(accessPoint.PublicAccessBlock)})
	d.Set("regions", flattenRegionReports(accessPoint.Regions))
	d.Set(names.AttrStatus, accessPoint.Status)

	return diags
}
