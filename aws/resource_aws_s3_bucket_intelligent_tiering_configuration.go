package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
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
				Required: true,
				ForceNew: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
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
							AtLeastOneOf: filterAtLeastOneOfKeys,
						},
						"tags": {
							Type:         schema.TypeMap,
							Optional:     true,
							AtLeastOneOf: filterAtLeastOneOfKeys,
							Elem:         &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"archive_configuration": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 2,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_tier": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(s3.IntelligentTieringAccessTier_Values(), false),
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
	config := d.Get("archive_configuration").(*schema.Set).List()
	id := d.Get("name").(string)

	log.Printf("[DEBUG] S3 bucket: %s, put policy: %s", bucket, id)

	// Set status from boolean value
	status := "Enabled"
	if d.Get("enabled").(bool) == false {
		status = "Disabled"
	}

	input := &s3.PutBucketIntelligentTieringConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(id),
		IntelligentTieringConfiguration: &s3.IntelligentTieringConfiguration{
			Filter:   expandS3IntelligentTieringFilter(d.Get("filter").([]interface{})),
			Id:       aws.String(id),
			Status:   aws.String(status),
			Tierings: expandS3IntelligentTieringConfigurations(config),
		},
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := s3conn.PutBucketIntelligentTieringConfiguration(input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = s3conn.PutBucketIntelligentTieringConfiguration(input)
	}

	d.SetId(bucket)

	return resourceAwsS3IntelligentTieringConfigurationRead(d, meta)
}

func expandS3IntelligentTieringFilter(l []interface{}) *s3.IntelligentTieringFilter {
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
		tags = keyvaluetags.New(v).IgnoreAws().S3Tags()
	}

	if prefix == "" && len(tags) == 0 {
		return nil
	}
	intelligentTieringFilter := &s3.IntelligentTieringFilter{}
	if prefix != "" && len(tags) > 0 {
		intelligentTieringFilter.And = &s3.IntelligentTieringAndOperator{
			Prefix: aws.String(prefix),
			Tags:   tags,
		}
	} else if len(tags) > 1 {
		intelligentTieringFilter.And = &s3.IntelligentTieringAndOperator{
			Tags: tags,
		}
	} else if len(tags) == 1 {
		intelligentTieringFilter.Tag = tags[0]
	} else {
		intelligentTieringFilter.Prefix = aws.String(prefix)
	}
	return intelligentTieringFilter
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
		if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") || isAWSErr(err, "NoSuchConfiguration", "The specified configuration does not exist.") {
			return nil
		}
		return fmt.Errorf("Error deleting Intelligent Tiering Configuration: %s", err)
	}

	return nil
}
