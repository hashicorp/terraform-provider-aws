// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKDataSource("aws_s3control_multi_region_access_point")
func dataSourceMultiRegionAccessPoint() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceMultiRegionAccessPointBlockRead,

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
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
						"bucket": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceMultiRegionAccessPointBlockRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, err := ConnForMRAP(ctx, meta.(*conns.AWSClient))

	if err != nil {
		return diag.FromErr(err)
	}

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}
	name := d.Get("name").(string)

	accessPoint, err := FindMultiRegionAccessPointByTwoPartKey(ctx, conn, accountID, name)

	if err != nil {
		return diag.Errorf("reading S3 Multi Region Access Point (%s): %s", name, err)
	}

	d.SetId(MultiRegionAccessPointCreateResourceID(accountID, name))

	alias := aws.StringValue(accessPoint.Alias)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "s3",
		AccountID: accountID,
		Resource:  fmt.Sprintf("accesspoint/%s", alias),
	}.String()
	d.Set("account_id", accountID)
	d.Set("alias", alias)
	d.Set("arn", arn)
	d.Set("created_at", aws.TimeValue(accessPoint.CreatedAt).Format(time.RFC3339))
	// https://docs.aws.amazon.com/AmazonS3/latest/userguide//MultiRegionAccessPointRequests.html#MultiRegionAccessPointHostnames.
	d.Set("domain_name", meta.(*conns.AWSClient).PartitionHostname(fmt.Sprintf("%s.accesspoint.s3-global", alias)))
	d.Set("name", accessPoint.Name)
	d.Set("public_access_block", []interface{}{flattenPublicAccessBlockConfiguration(accessPoint.PublicAccessBlock)})
	d.Set("regions", flattenRegionReports(accessPoint.Regions))
	d.Set("status", accessPoint.Status)

	return nil
}
