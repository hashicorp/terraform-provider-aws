package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsCloudFrontDistribution() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCloudfrontDistributionRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"in_progress_validation_batches": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"active_trusted_signers": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsCloudfrontDistributionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn
	id := d.Get("id").(string)
	input := &cloudfront.GetDistributionInput{
		Id: aws.String(id),
	}

	log.Printf("[DEBUG] Reading CloudFront Distribution: %s", input)
	out, err := conn.GetDistribution(input)
	if err != nil {
		return fmt.Errorf("Failed getting CloudFront distribution (%s): %s", id, err)
	}
	dist := out.Distribution
	etag := out.ETag
	d.SetId(*dist.Id)

	d.Set("arn", dist.ARN)
	d.Set("status", dist.Status)
	d.Set("domain_name", dist.DomainName)
	d.Set("last_modified_time", dist.LastModifiedTime)
	d.Set("in_progress_validation_batches", dist.InProgressInvalidationBatches)
	d.Set("etag", etag)

	err = d.Set("active_trusted_signers", flattenActiveTrustedSigners(dist.ActiveTrustedSigners))
	if err != nil {
		return err
	}

	return nil
}
