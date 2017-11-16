package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsS3Bucket() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsS3BucketRead,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"server_side_encryption_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"rule": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"apply_server_side_encryption_by_default": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"kms_master_key_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"sse_algorithm": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bucket_domain_name": {
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

func dataSourceAwsS3BucketRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)

	input := &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}

	log.Printf("[DEBUG] Reading S3 bucket: %s", input)
	_, err := conn.HeadBucket(input)

	if err != nil {
		return fmt.Errorf("Failed getting S3 bucket: %s Bucket: %q", err, bucket)
	}

	d.SetId(bucket)
	d.Set("arn", fmt.Sprintf("arn:%s:s3:::%s", meta.(*AWSClient).partition, bucket))
	d.Set("bucket_domain_name", bucketDomainName(bucket))

	if err := bucketLocation(d, bucket, conn); err != nil {
		return err
	}

	if err := bucketEncryption(d, bucket, conn); err != nil {
		return err
	}

	return nil
}

func bucketEncryption(data *schema.ResourceData, bucket string, conn *s3.S3) error {
	input := &s3.GetBucketEncryptionInput{
		Bucket: aws.String(bucket),
	}
	output, err := conn.GetBucketEncryption(input)
	if err != nil {
		if isAWSErr(err, "ServerSideEncryptionConfigurationNotFoundError", "encryption configuration was not found") {
			log.Printf("[DEBUG] Default encryption is not enabled for %s", bucket)
			data.Set("server_side_encryption_configuration", []map[string]interface{}{
				{
					"enabled": false,
				},
			})
			return nil
		} else {
			return err
		}
	}
	if ruleCount := len(output.ServerSideEncryptionConfiguration.Rules); ruleCount != 1 {
		return fmt.Errorf("expect one rule returned but there are %d rules. Changes required in the data source to support this", ruleCount)
	}
	defaultRuleConfiguration := output.ServerSideEncryptionConfiguration.Rules[0].ApplyServerSideEncryptionByDefault
	defaultRule := make([]map[string]interface{}, 1)
	defaultRule[0]["kms_master_key_id"] = aws.StringValue(defaultRuleConfiguration.KMSMasterKeyID)
	defaultRule[0]["sse_algorithm"] = aws.StringValue(defaultRuleConfiguration.SSEAlgorithm)

	encryptionConfiguration := make([]map[string]interface{}, 1)
	encryptionConfiguration[0]["enabled"] = true
	encryptionConfiguration[0]["rule"] = make([]map[string]interface{}, 1)
	encryptionConfiguration[0]["rule"].(map[string]interface{})["apply_server_side_encryption_by_default"] = defaultRule

	data.Set("server_side_encryption_configuration", encryptionConfiguration)
	return nil
}

func bucketLocation(d *schema.ResourceData, bucket string, conn *s3.S3) error {
	location, err := conn.GetBucketLocation(
		&s3.GetBucketLocationInput{
			Bucket: aws.String(bucket),
		},
	)
	if err != nil {
		return err
	}
	var region string
	if location.LocationConstraint != nil {
		region = *location.LocationConstraint
	}
	region = normalizeRegion(region)
	if err := d.Set("region", region); err != nil {
		return err
	}

	hostedZoneID := HostedZoneIDForRegion(region)
	if err := d.Set("hosted_zone_id", hostedZoneID); err != nil {
		return err
	}

	_, websiteErr := conn.GetBucketWebsite(
		&s3.GetBucketWebsiteInput{
			Bucket: aws.String(bucket),
		},
	)

	if websiteErr == nil {
		websiteEndpoint := WebsiteEndpoint(bucket, region)
		if err := d.Set("website_endpoint", websiteEndpoint.Endpoint); err != nil {
			return err
		}
		if err := d.Set("website_domain", websiteEndpoint.Domain); err != nil {
			return err
		}
	}
	return nil
}
