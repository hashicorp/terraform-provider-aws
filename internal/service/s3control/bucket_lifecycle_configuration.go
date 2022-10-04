package s3control

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceBucketLifecycleConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceBucketLifecycleConfigurationCreate,
		Read:   resourceBucketLifecycleConfigurationRead,
		Update: resourceBucketLifecycleConfigurationUpdate,
		Delete: resourceBucketLifecycleConfigurationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"rule": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"abort_incomplete_multipart_upload": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"days_after_initiation": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
						"expiration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"date": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
											value := v.(string)

											_, err := time.Parse("2006-01-02", value)

											if err != nil {
												errors = append(errors, fmt.Errorf("%q should be in YYYY-MM-DD date format", value))
											}

											return
										},
									},
									"days": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"expired_object_delete_marker": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false, // Prevent SDK TypeSet difference issues
									},
								},
							},
						},
						"filter": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"prefix": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"tags": tftags.TagsSchema(),
								},
							},
						},
						"id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"status": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  s3control.ExpirationStatusEnabled,
						},
					},
				},
			},
		},
	}
}

func resourceBucketLifecycleConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3ControlConn

	bucket := d.Get("bucket").(string)

	parsedArn, err := arn.Parse(bucket)

	if err != nil {
		return fmt.Errorf("error parsing S3 Control Bucket ARN (%s): %w", bucket, err)
	}

	if parsedArn.AccountID == "" {
		return fmt.Errorf("error parsing S3 Control Bucket ARN (%s): unknown format", d.Id())
	}

	input := &s3control.PutBucketLifecycleConfigurationInput{
		AccountId: aws.String(parsedArn.AccountID),
		Bucket:    aws.String(bucket),
		LifecycleConfiguration: &s3control.LifecycleConfiguration{
			Rules: expandLifecycleRules(d.Get("rule").(*schema.Set).List()),
		},
	}

	_, err = conn.PutBucketLifecycleConfiguration(input)

	if err != nil {
		return fmt.Errorf("error creating S3 Control Lifecycle Configuration (%s): %w", bucket, err)
	}

	d.SetId(bucket)

	return resourceBucketLifecycleConfigurationRead(d, meta)
}

func resourceBucketLifecycleConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3ControlConn

	parsedArn, err := arn.Parse(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing S3 Control Bucket ARN (%s): %w", d.Id(), err)
	}

	if parsedArn.AccountID == "" {
		return fmt.Errorf("error parsing S3 Control Bucket ARN (%s): unknown format", d.Id())
	}

	input := &s3control.GetBucketLifecycleConfigurationInput{
		AccountId: aws.String(parsedArn.AccountID),
		Bucket:    aws.String(d.Id()),
	}

	output, err := conn.GetBucketLifecycleConfiguration(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, "NoSuchBucket") {
		log.Printf("[WARN] S3 Control Lifecycle Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, "NoSuchLifecycleConfiguration") {
		log.Printf("[WARN] S3 Control Lifecycle Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, "NoSuchOutpost") {
		log.Printf("[WARN] S3 Control Lifecycle Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading S3 Control Lifecycle Configuration (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error reading S3 Control Lifecycle Configuration (%s): empty response", d.Id())
	}

	d.Set("bucket", d.Id())

	if err := d.Set("rule", flattenLifecycleRules(output.Rules)); err != nil {
		return fmt.Errorf("error setting rule: %w", err)
	}

	return nil
}

func resourceBucketLifecycleConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3ControlConn

	parsedArn, err := arn.Parse(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing S3 Control Bucket ARN (%s): %w", d.Id(), err)
	}

	if parsedArn.AccountID == "" {
		return fmt.Errorf("error parsing S3 Control Bucket ARN (%s): unknown format", d.Id())
	}

	input := &s3control.PutBucketLifecycleConfigurationInput{
		AccountId: aws.String(parsedArn.AccountID),
		Bucket:    aws.String(d.Id()),
		LifecycleConfiguration: &s3control.LifecycleConfiguration{
			Rules: expandLifecycleRules(d.Get("rule").(*schema.Set).List()),
		},
	}

	_, err = conn.PutBucketLifecycleConfiguration(input)

	if err != nil {
		return fmt.Errorf("error updating S3 Control Lifecycle Configuration (%s): %w", d.Id(), err)
	}

	return resourceBucketLifecycleConfigurationRead(d, meta)
}

func resourceBucketLifecycleConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3ControlConn

	parsedArn, err := arn.Parse(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing S3 Control Bucket ARN (%s): %w", d.Id(), err)
	}

	if parsedArn.AccountID == "" {
		return fmt.Errorf("error parsing S3 Control Bucket ARN (%s): unknown format", d.Id())
	}

	input := &s3control.DeleteBucketLifecycleConfigurationInput{
		AccountId: aws.String(parsedArn.AccountID),
		Bucket:    aws.String(d.Id()),
	}

	_, err = conn.DeleteBucketLifecycleConfiguration(input)

	if tfawserr.ErrCodeEquals(err, "NoSuchBucket") {
		return nil
	}

	if tfawserr.ErrCodeEquals(err, "NoSuchLifecycleConfiguration") {
		return nil
	}

	if tfawserr.ErrCodeEquals(err, "NoSuchOutpost") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting S3 Control Lifecycle Configuration (%s): %w", d.Id(), err)
	}

	return nil
}

func expandAbortIncompleteMultipartUpload(tfList []interface{}) *s3control.AbortIncompleteMultipartUpload {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &s3control.AbortIncompleteMultipartUpload{}

	if v, ok := tfMap["days_after_initiation"].(int); ok && v != 0 {
		apiObject.DaysAfterInitiation = aws.Int64(int64(v))
	}

	return apiObject
}

func expandLifecycleExpiration(tfList []interface{}) *s3control.LifecycleExpiration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &s3control.LifecycleExpiration{}

	if v, ok := tfMap["date"].(string); ok && v != "" {
		parsedDate, err := time.Parse("2006-01-02", v)

		if err == nil {
			apiObject.Date = aws.Time(parsedDate)
		}
	}

	if v, ok := tfMap["days"].(int); ok && v != 0 {
		apiObject.Days = aws.Int64(int64(v))
	}

	if v, ok := tfMap["expired_object_delete_marker"].(bool); ok && v {
		apiObject.ExpiredObjectDeleteMarker = aws.Bool(v)
	}

	return apiObject
}

func expandLifecycleRules(tfList []interface{}) []*s3control.LifecycleRule {
	var apiObjects []*s3control.LifecycleRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandLifecycleRule(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandLifecycleRule(tfMap map[string]interface{}) *s3control.LifecycleRule {
	if len(tfMap) == 0 {
		return nil
	}

	apiObject := &s3control.LifecycleRule{}

	if v, ok := tfMap["abort_incomplete_multipart_upload"].([]interface{}); ok && len(v) > 0 {
		apiObject.AbortIncompleteMultipartUpload = expandAbortIncompleteMultipartUpload(v)
	}

	if v, ok := tfMap["expiration"].([]interface{}); ok && len(v) > 0 {
		apiObject.Expiration = expandLifecycleExpiration(v)
	}

	if v, ok := tfMap["filter"].([]interface{}); ok && len(v) > 0 {
		apiObject.Filter = expandLifecycleRuleFilter(v)
	}

	if v, ok := tfMap["id"].(string); ok && v != "" {
		apiObject.ID = aws.String(v)
	}

	if v, ok := tfMap["status"].(string); ok && v != "" {
		apiObject.Status = aws.String(v)
	}

	// Terraform Plugin SDK sometimes sends map with only empty configuration blocks:
	//   map[abort_incomplete_multipart_upload:[] expiration:[] filter:[] id: status:]
	// This is to prevent this error: InvalidParameter: 1 validation error(s) found.
	//  - missing required field, PutBucketLifecycleConfigurationInput.LifecycleConfiguration.Rules[0].Status.
	if apiObject.ID == nil && apiObject.Status == nil {
		return nil
	}

	return apiObject
}

func expandLifecycleRuleFilter(tfList []interface{}) *s3control.LifecycleRuleFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &s3control.LifecycleRuleFilter{}

	if v, ok := tfMap["prefix"].(string); ok && v != "" {
		apiObject.Prefix = aws.String(v)
	}

	if v, ok := tfMap["tags"].(map[string]interface{}); ok && len(v) > 0 {
		// See also aws_s3_bucket ReplicationRule.Filter handling
		if len(v) == 1 {
			apiObject.Tag = Tags(tftags.New(v))[0]
		} else {
			apiObject.And = &s3control.LifecycleRuleAndOperator{
				Prefix: apiObject.Prefix,
				Tags:   Tags(tftags.New(v)),
			}
			apiObject.Prefix = nil
		}
	}

	return apiObject
}

func flattenAbortIncompleteMultipartUpload(apiObject *s3control.AbortIncompleteMultipartUpload) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DaysAfterInitiation; v != nil {
		tfMap["days_after_initiation"] = aws.Int64Value(v)
	}

	return []interface{}{tfMap}
}

func flattenLifecycleExpiration(apiObject *s3control.LifecycleExpiration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Date; v != nil {
		tfMap["date"] = aws.TimeValue(v).Format("2006-01-02")
	}

	if v := apiObject.Days; v != nil {
		tfMap["days"] = aws.Int64Value(v)
	}

	if v := apiObject.ExpiredObjectDeleteMarker; v != nil {
		tfMap["expired_object_delete_marker"] = aws.BoolValue(v)
	}

	return []interface{}{tfMap}
}

func flattenLifecycleRules(apiObjects []*s3control.LifecycleRule) []interface{} {
	var tfMaps []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfMaps = append(tfMaps, flattenLifecycleRule(apiObject))
	}

	return tfMaps
}

func flattenLifecycleRule(apiObject *s3control.LifecycleRule) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AbortIncompleteMultipartUpload; v != nil {
		tfMap["abort_incomplete_multipart_upload"] = flattenAbortIncompleteMultipartUpload(v)
	}

	if v := apiObject.Expiration; v != nil {
		tfMap["expiration"] = flattenLifecycleExpiration(v)
	}

	if v := apiObject.Filter; v != nil {
		tfMap["filter"] = flattenLifecycleRuleFilter(v)
	}

	if v := apiObject.ID; v != nil {
		tfMap["id"] = aws.StringValue(v)
	}

	if v := apiObject.Status; v != nil {
		tfMap["status"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenLifecycleRuleFilter(apiObject *s3control.LifecycleRuleFilter) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.And != nil {
		if v := apiObject.And.Prefix; v != nil {
			tfMap["prefix"] = aws.StringValue(v)
		}

		if v := apiObject.And.Tags; v != nil {
			tfMap["tags"] = KeyValueTags(v).IgnoreAWS().Map()
		}
	} else {
		if v := apiObject.Prefix; v != nil {
			tfMap["prefix"] = aws.StringValue(v)
		}

		if v := apiObject.Tag; v != nil {
			tfMap["tags"] = KeyValueTags([]*s3control.S3Tag{v}).IgnoreAWS().Map()
		}
	}

	return []interface{}{tfMap}
}
