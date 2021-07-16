package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	// "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsS3IntelligentTieringConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsS3IntelligentTieringConfigurationPut,
		Read:   resourceAwsS3IntelligentTieringConfigurationRead,
		Update: resourceAwsS3IntelligentTieringConfigurationPut,
		Delete: resourceAwsS3IntelligentTieringConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tagsSchema(),
			"archive_configuration": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 2,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_tier": {
							Type:     schema.TypeString,
							Required: true,
						},
						"days": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsS3IntelligentTieringConfigurationPut(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)
	config := d.Get("tiering_configuration").(*schema.Set).List()

	id := d.Get("name").(string)

	log.Printf("[DEBUG] S3 bucket: %s, put policy: %s", bucket, id)

	params := &s3.PutBucketIntelligentTieringConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(id),
		IntelligentTieringConfiguration: &s3.IntelligentTieringConfiguration{
			Filter: &s3.IntelligentTieringFilter{
				And: &s3.IntelligentTieringAndOperator{
					Prefix: aws.String("test"),
				},
			},
			Id:       aws.String(id),
			Status:   aws.String("Enabled"),
			Tierings: expandS3IntelligentTieringConfigurations(config),
		},
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := s3conn.PutBucketIntelligentTieringConfiguration(params)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = s3conn.PutBucketIntelligentTieringConfiguration(params)
	}
	if err != nil {
		return fmt.Errorf("Error putting Intelligent Tiering Configuration: %s", err)
	}

	d.SetId(bucket)

	return resourceAwsS3IntelligentTieringConfigurationRead(d, meta)
}

func expandS3IntelligentTieringConfigurations(tfList []interface{}) []*s3.Tiering {
	var apiObjects []*s3.Tiering

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandS3IntelligentTieringConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandS3IntelligentTieringConfiguration(tfMap map[string]interface{}) *s3.Tiering {
	if len(tfMap) == 0 {
		return nil
	}

	apiObject := &s3.Tiering{}

	log.Printf("[DEBUG] This is the map: %s", tfMap)

	if v, ok := tfMap["access_tier"].(string); ok && v != "" {
		apiObject.AccessTier = aws.String(v)
	}

	if v, ok := tfMap["days"].(int); ok && v != 0 {
		apiObject.Days = aws.Int64(int64(v))
	}

	log.Printf("[DEBUG] This is the apiObject: %s", apiObject)

	return apiObject
}

func resourceAwsS3IntelligentTieringConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)
	id := d.Get("name").(string)

	log.Printf("[DEBUG] S3 bucket policy, read for bucket: %s", bucket)
	s3conn.GetBucketIntelligentTieringConfiguration(&s3.GetBucketIntelligentTieringConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(id),
	})

	// v := ""
	// if err == nil && pol.IntelligentTieringConfiguration != nil {
	// 	v = pol.IntelligentTieringConfiguration
	// }
	// if err := d.Set("id", v); err != nil {
	// 	return err
	// }
	// if err := d.Set("bucket", d.Id()); err != nil {
	// 	return err
	// }

	return nil
}

func resourceAwsS3IntelligentTieringConfigurationDelete(d *schema.ResourceData, meta interface{}) error {

	s3conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)
	id := d.Get("name").(string)

	log.Printf("[DEBUG] S3 bucket: %s, delete policy", bucket)
	_, err := s3conn.DeleteBucketIntelligentTieringConfiguration(&s3.DeleteBucketIntelligentTieringConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(id),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoSuchBucket" {
			return nil
		}
		return fmt.Errorf("Error deleting Intelligent Tiering Configuration: %s", err)
	}

	return nil
}
