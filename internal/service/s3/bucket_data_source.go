package s3

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceBucket() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBucketRead,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
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
			"website_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"website_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceBucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn()

	bucket := d.Get("bucket").(string)

	input := &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}

	log.Printf("[DEBUG] Reading S3 bucket: %s", input)
	_, err := conn.HeadBucketWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Failed getting S3 bucket (%s): %s", bucket, err)
	}

	d.SetId(bucket)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "s3",
		Resource:  bucket,
	}.String()
	d.Set("arn", arn)
	d.Set("bucket_domain_name", meta.(*conns.AWSClient).PartitionHostname(fmt.Sprintf("%s.s3", bucket)))

	err = bucketLocation(ctx, meta.(*conns.AWSClient), d, bucket)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting S3 Bucket location: %s", err)
	}

	regionalDomainName, err := BucketRegionalDomainName(bucket, d.Get("region").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting S3 Bucket regional domain name: %s", err)
	}
	d.Set("bucket_regional_domain_name", regionalDomainName)

	return diags
}

func bucketLocation(ctx context.Context, client *conns.AWSClient, d *schema.ResourceData, bucket string) error {
	region, err := s3manager.GetBucketRegionWithClient(ctx, client.S3Conn(), bucket, func(r *request.Request) {
		// By default, GetBucketRegion forces virtual host addressing, which
		// is not compatible with many non-AWS implementations. Instead, pass
		// the provider s3_force_path_style configuration, which defaults to
		// false, but allows override.
		r.Config.S3ForcePathStyle = client.S3Conn().Config.S3ForcePathStyle

		// By default, GetBucketRegion uses anonymous credentials when doing
		// a HEAD request to get the bucket region. This breaks in aws-cn regions
		// when the account doesn't have an ICP license to host public content.
		// Use the current credentials when getting the bucket region.
		r.Config.Credentials = client.S3Conn().Config.Credentials
	})
	if err != nil {
		return err
	}
	if err := d.Set("region", region); err != nil {
		return err
	}

	hostedZoneID, err := HostedZoneIDForRegion(region)
	if err != nil {
		log.Printf("[WARN] %s", err)
	} else {
		d.Set("hosted_zone_id", hostedZoneID)
	}

	_, websiteErr := client.S3Conn().GetBucketWebsite(
		&s3.GetBucketWebsiteInput{
			Bucket: aws.String(bucket),
		},
	)

	if websiteErr == nil {
		websiteEndpoint := WebsiteEndpoint(client, bucket, region)
		if err := d.Set("website_endpoint", websiteEndpoint.Endpoint); err != nil {
			return err
		}
		if err := d.Set("website_domain", websiteEndpoint.Domain); err != nil {
			return err
		}
	}
	return nil
}
