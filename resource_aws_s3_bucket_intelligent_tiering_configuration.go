package aws

import (
	"fmt"
	"log"
	"strings"
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
			"tier": {
				Type:     schema.TypeSet,
				Optional: true,
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
	tiers := d.Get("tier").(*schema.Set).List()
	name := d.Get("name").(string)

	log.Printf("[DEBUG] S3 bucket: %s, put intelligent tiering configuration: %s", bucket, name)

	status := resourceAwsS3BucketIntelligentTieringConfigurationParseEnabled(d.Get("enabled").(bool))

	input := &s3.PutBucketIntelligentTieringConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
		IntelligentTieringConfiguration: &s3.IntelligentTieringConfiguration{
			Filter:   expandS3IntelligentTieringFilter(d.Get("filter").([]interface{})),
			Id:       aws.String(name),
			Status:   status,
			Tierings: expandS3IntelligentTieringConfigurations(tiers),
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

	d.SetId(fmt.Sprintf("%s:%s", bucket, name))

	return resourceAwsS3IntelligentTieringConfigurationRead(d, meta)
}

func resourceAwsS3IntelligentTieringConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn

	bucket, name, err := resourceAwsS3BucketIntelligentTieringConfigurationParseID(d.Id())
	if err != nil {
		return err
	}

	d.Set("bucket", bucket)
	d.Set("name", name)

	input := &s3.GetBucketIntelligentTieringConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	}
	log.Printf("[DEBUG] Reading S3 bucket Intelligent Tiering Configuration: %s, with id %s", input, d.Id())

	output, err := s3conn.GetBucketIntelligentTieringConfiguration(input)

	if !d.IsNewResource() && isAWSErr(err, s3.ErrCodeNoSuchBucket, "") {
		log.Printf("[WARN] S3 Bucket Intelligent Tiering Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if !d.IsNewResource() && isAWSErr(err, "NoSuchConfiguration", "The specified configuration does not exist.") {
		log.Printf("[WARN] S3 Bucket Intelligent Tiering Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("here error getting S3 Bucket Intelligent Tiering Configuration (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting S3 Bucket Intelligent Tiering Configuration (%s): empty response", d.Id())
	}

	if err := d.Set("filter", flattenS3IntelligentTieringFilter(output.IntelligentTieringConfiguration.Filter)); err != nil {
		return fmt.Errorf("error setting filter: %w", err)
	}

	if err := d.Set("enabled", resourceAwsS3BucketIntelligentTieringConfigurationParseStatus(output.IntelligentTieringConfiguration.Status)); err != nil {
		return fmt.Errorf("error setting enabled status: %w", err)
	}

	if err = d.Set("tier", flattenS3IntelligentTieringConfiguration(output.IntelligentTieringConfiguration.Tierings)); err != nil {
		return fmt.Errorf("error setting archive tier: %w", err)
	}

	return nil
}

func resourceAwsS3IntelligentTieringConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn

	bucket, name, err := resourceAwsS3BucketIntelligentTieringConfigurationParseID(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] S3 bucket: %s, delete intelligent tiering configuration", bucket)

	_, err = s3conn.DeleteBucketIntelligentTieringConfiguration(&s3.DeleteBucketIntelligentTieringConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	})

	if err != nil {
		if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") || isAWSErr(err, "NoSuchConfiguration", "The specified configuration does not exist.") {
			return nil
		}
		return fmt.Errorf("Error deleting Intelligent Tiering Configuration: %s", err)
	}

	return nil
}

func resourceAwsS3BucketIntelligentTieringConfigurationParseID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("please make sure the ID is in the form BUCKET:NAME (i.e. my-bucket:EntireBucket")
	}
	bucket := idParts[0]
	name := idParts[1]
	return bucket, name, nil
}

func waitForDeleteS3BucketIntelligentTieringConfiguration(conn *s3.S3, bucket, name string, timeout time.Duration) error {
	input := &s3.GetBucketIntelligentTieringConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	}

	err := resource.Retry(timeout, func() *resource.RetryError {

		output, err := conn.GetBucketIntelligentTieringConfiguration(input)

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if output != nil && output.IntelligentTieringConfiguration != nil {
			return resource.RetryableError(fmt.Errorf("S3 bucket intelligent tiering configuration exists: %v", output))
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.GetBucketIntelligentTieringConfiguration(input)
	}

	if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") || isAWSErr(err, "NoSuchConfiguration", "The specified configuration does not exist.") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting S3 Bucket Intelligent Tiering Configuration \"%s\": %w", name, err)
	}

	return nil
}

func resourceAwsS3BucketIntelligentTieringConfigurationParseStatus(status *string) bool {
	if status == nil {
		return false
	}

	if strings.ToLower(*status) == "enabled" {
		return true
	}

	return false
}

func resourceAwsS3BucketIntelligentTieringConfigurationParseEnabled(enabled bool) *string {
	if enabled == false {
		return aws.String("Disabled")
	}

	return aws.String("Enabled")
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

	if v, ok := tfMap["access_tier"].(string); ok && v != "" {
		apiObject.AccessTier = aws.String(v)
	}

	if v, ok := tfMap["days"].(int); ok && v != 0 {
		apiObject.Days = aws.Int64(int64(v))
	}

	return apiObject
}

func flattenS3IntelligentTieringFilter(intelligentTieringFilter *s3.IntelligentTieringFilter) []map[string]interface{} {
	if intelligentTieringFilter == nil {
		return nil
	}

	result := make(map[string]interface{})
	if intelligentTieringFilter.And != nil {
		and := *intelligentTieringFilter.And
		if and.Prefix != nil {
			result["prefix"] = *and.Prefix
		}
		if and.Tags != nil {
			result["tags"] = keyvaluetags.S3KeyValueTags(and.Tags).IgnoreAws().Map()
		}
	} else if intelligentTieringFilter.Prefix != nil {
		result["prefix"] = *intelligentTieringFilter.Prefix
	} else if intelligentTieringFilter.Tag != nil {
		tags := []*s3.Tag{
			intelligentTieringFilter.Tag,
		}
		result["tags"] = keyvaluetags.S3KeyValueTags(tags).IgnoreAws().Map()
	} else {
		return nil
	}
	return []map[string]interface{}{result}
}

func flattenS3IntelligentTieringConfiguration(tieringConfigurations []*s3.Tiering) []map[string]interface{} {
	if tieringConfigurations == nil {
		return []map[string]interface{}{}
	}

	ac := make([]map[string]interface{}, 0, len(tieringConfigurations))
	for _, c := range tieringConfigurations {
		q := make(map[string]interface{})
		if c.AccessTier != nil {
			q["access_tier"] = aws.StringValue(c.AccessTier)
		}

		if c.Days != nil {
			q["days"] = int(aws.Int64Value(c.Days))
		}
		ac = append(ac, q)
	}

	return ac
}
