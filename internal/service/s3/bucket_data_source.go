// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_s3_bucket")
func DataSourceBucket() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBucketRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"bucket_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bucket_regional_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hosted_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"website_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"website_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceBucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	awsClient := meta.(*conns.AWSClient)
	conn := awsClient.S3Client(ctx)

	bucket := d.Get("bucket").(string)
	err := findBucket(ctx, conn, bucket)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s): %s", bucket, err)
	}

	region, err := manager.GetBucketRegion(ctx, conn, bucket,
		func(o *s3.Options) {
			// By default, GetBucketRegion forces virtual host addressing, which
			// is not compatible with many non-AWS implementations. Instead, pass
			// the provider s3_force_path_style configuration, which defaults to
			// false, but allows override.
			o.UsePathStyle = awsClient.S3UsePathStyle()
		},
		func(o *s3.Options) {
			// By default, GetBucketRegion uses anonymous credentials when doing
			// a HEAD request to get the bucket region. This breaks in aws-cn regions
			// when the account doesn't have an ICP license to host public content.
			// Use the current credentials when getting the bucket region.
			o.Credentials = awsClient.CredentialsProvider()
		})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s) Region: %s", bucket, err)
	}

	d.SetId(bucket)
	arn := arn.ARN{
		Partition: awsClient.Partition,
		Service:   "s3",
		Resource:  bucket,
	}.String()
	d.Set("arn", arn)
	d.Set("bucket_domain_name", awsClient.PartitionHostname(fmt.Sprintf("%s.s3", bucket)))
	if regionalDomainName, err := BucketRegionalDomainName(bucket, region); err == nil {
		d.Set("bucket_regional_domain_name", regionalDomainName)
	} else {
		log.Printf("[WARN] BucketRegionalDomainName: %s", err)
	}
	if hostedZoneID, err := HostedZoneIDForRegion(region); err == nil {
		d.Set("hosted_zone_id", hostedZoneID)
	} else {
		log.Printf("[WARN] HostedZoneIDForRegion: %s", err)
	}
	d.Set("region", region)
	if _, err := findBucketWebsite(ctx, conn, bucket, ""); err == nil {
		website := WebsiteEndpoint(awsClient, bucket, region)
		d.Set("website_domain", website.Domain)
		d.Set("website_endpoint", website.Endpoint)
	} else if !tfresource.NotFound(err) {
		log.Printf("[WARN] Reading S3 Bucket (%s) Website: %s", bucket, err)
	}

	return diags
}
