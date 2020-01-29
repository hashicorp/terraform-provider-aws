package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsS3BucketAnalyticsConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsS3BucketAnalyticsConfigurationPut,
		Read:   resourceAwsS3BucketAnalyticsConfigurationRead,
		Update: resourceAwsS3BucketAnalyticsConfigurationPut,
		Delete: resourceAwsS3BucketAnalyticsConfigurationDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prefix": {
							Type:         schema.TypeString,
							Optional:     true,
							AtLeastOneOf: []string{"filter.0.prefix", "filter.0.tags"},
						},
						"tags": schemaWithAtLeastOneOf(
							tagsSchema(),
							[]string{"filter.0.prefix", "filter.0.tags"},
						),
					},
				},
			},
			"storage_class_analysis": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func schemaWithAtLeastOneOf(schema *schema.Schema, keys []string) *schema.Schema {
	schema.AtLeastOneOf = keys
	return schema
}

func resourceAwsS3BucketAnalyticsConfigurationPut(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)
	name := d.Get("name").(string)

	log.Printf("[DEBUG] S3 bucket %q, add analytics configuration %q", bucket, name)

	storageClassAnalysis := &s3.StorageClassAnalysis{}
	analyticsConfiguration := &s3.AnalyticsConfiguration{
		Id:                   aws.String(name),
		Filter:               expandS3AnalyticsFilter(d.Get("filter").([]interface{})),
		StorageClassAnalysis: storageClassAnalysis,
	}

	input := &s3.PutBucketAnalyticsConfigurationInput{
		Bucket:                 aws.String(bucket),
		Id:                     aws.String(name),
		AnalyticsConfiguration: analyticsConfiguration,
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := s3conn.PutBucketAnalyticsConfiguration(input)
		if err != nil {
			if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = s3conn.PutBucketAnalyticsConfiguration(input)
	}
	if err != nil {
		return fmt.Errorf("Error adding S3 analytics configuration: %w", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", bucket, name))

	return resourceAwsS3BucketAnalyticsConfigurationRead(d, meta)
}

func resourceAwsS3BucketAnalyticsConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	bucket, name, err := resourceAwsS3BucketAnalyticsConfigurationParseID(d.Id())
	if err != nil {
		return err
	}

	d.Set("bucket", bucket)
	d.Set("name", name)

	input := &s3.GetBucketAnalyticsConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	}

	log.Printf("[DEBUG] Reading S3 bucket analytics configuration: %s", input)
	output, err := conn.GetBucketAnalyticsConfiguration(input)
	if err != nil {
		if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") || isAWSErr(err, "NoSuchConfiguration", "The specified configuration does not exist.") {
			log.Printf("[WARN] %s S3 bucket analytics configuration not found, removing from state.", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if err := d.Set("filter", flattenS3AnalyticsFilter(output.AnalyticsConfiguration.Filter)); err != nil {
		return err
	}

	return nil
}

func resourceAwsS3BucketAnalyticsConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	bucket, name, err := resourceAwsS3BucketAnalyticsConfigurationParseID(d.Id())
	if err != nil {
		return err
	}

	input := &s3.DeleteBucketAnalyticsConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	}

	log.Printf("[DEBUG] Deleting S3 bucket analytics configuration: %s", input)
	_, err = conn.DeleteBucketAnalyticsConfiguration(input)
	if err != nil {
		if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") || isAWSErr(err, "NoSuchConfiguration", "The specified configuration does not exist.") {
			return nil
		}
		return fmt.Errorf("Error deleting S3 analytics configuration: %w", err)
	}

	return waitForDeleteS3BucketAnalyticsConfiguration(conn, bucket, name, 1*time.Minute)
}

func resourceAwsS3BucketAnalyticsConfigurationParseID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("please make sure the ID is in the form BUCKET:NAME (i.e. my-bucket:EntireBucket")
	}
	bucket := idParts[0]
	name := idParts[1]
	return bucket, name, nil
}

func expandS3AnalyticsFilter(l []interface{}) *s3.AnalyticsFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	var prefix string
	if v, ok := m["prefix"]; ok {
		prefix = v.(string)
	}

	var tags []*s3.Tag
	if v, ok := m["tags"]; ok {
		tags = tagsFromMapS3(v.(map[string]interface{}))
	}

	if prefix == "" && len(tags) == 0 {
		return nil
	}
	analyticsFilter := &s3.AnalyticsFilter{}
	if prefix != "" && len(tags) > 0 {
		analyticsFilter.And = &s3.AnalyticsAndOperator{
			Prefix: aws.String(prefix),
			Tags:   tags,
		}
	} else if len(tags) > 1 {
		analyticsFilter.And = &s3.AnalyticsAndOperator{
			Tags: tags,
		}
	} else if len(tags) == 1 {
		analyticsFilter.Tag = tags[0]
	} else {
		analyticsFilter.Prefix = aws.String(prefix)
	}
	return analyticsFilter
}

func flattenS3AnalyticsFilter(analyticsFilter *s3.AnalyticsFilter) []map[string]interface{} {
	if analyticsFilter == nil {
		return nil
	}

	result := make(map[string]interface{})
	if analyticsFilter.And != nil {
		and := *analyticsFilter.And
		if and.Prefix != nil {
			result["prefix"] = *and.Prefix
		}
		if and.Tags != nil {
			result["tags"] = tagsToMapS3(and.Tags)
		}
	} else if analyticsFilter.Prefix != nil {
		result["prefix"] = *analyticsFilter.Prefix
	} else if analyticsFilter.Tag != nil {
		tags := []*s3.Tag{
			analyticsFilter.Tag,
		}
		result["tags"] = tagsToMapS3(tags)
	} else {
		return nil
	}
	return []map[string]interface{}{result}
}

func waitForDeleteS3BucketAnalyticsConfiguration(conn *s3.S3, bucket, name string, timeout time.Duration) error {
	err := resource.Retry(timeout, func() *resource.RetryError {
		input := &s3.GetBucketAnalyticsConfigurationInput{
			Bucket: aws.String(bucket),
			Id:     aws.String(name),
		}
		log.Printf("[DEBUG] Reading S3 bucket analytics configuration: %s", input)
		output, err := conn.GetBucketAnalyticsConfiguration(input)
		if err != nil {
			if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") || isAWSErr(err, "NoSuchConfiguration", "The specified configuration does not exist.") {
				return nil
			}
			return resource.NonRetryableError(err)
		}
		if output.AnalyticsConfiguration != nil {
			return resource.RetryableError(fmt.Errorf("S3 bucket analytics configuration exists: %v", output))
		}

		return nil
	})

	return err
}
