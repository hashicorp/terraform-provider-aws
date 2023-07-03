// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_s3control_bucket_lifecycle_configuration")
func resourceBucketLifecycleConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketLifecycleConfigurationCreate,
		ReadWithoutTimeout:   resourceBucketLifecycleConfigurationRead,
		UpdateWithoutTimeout: resourceBucketLifecycleConfigurationUpdate,
		DeleteWithoutTimeout: resourceBucketLifecycleConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceBucketLifecycleConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn(ctx)

	bucket := d.Get("bucket").(string)

	parsedArn, err := arn.Parse(bucket)

	if err != nil {
		return diag.FromErr(err)
	}

	if parsedArn.AccountID == "" {
		return diag.Errorf("parsing S3 Control Bucket ARN (%s): unknown format", d.Id())
	}

	input := &s3control.PutBucketLifecycleConfigurationInput{
		AccountId: aws.String(parsedArn.AccountID),
		Bucket:    aws.String(bucket),
		LifecycleConfiguration: &s3control.LifecycleConfiguration{
			Rules: expandLifecycleRules(ctx, d.Get("rule").(*schema.Set).List()),
		},
	}

	_, err = conn.PutBucketLifecycleConfigurationWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating S3 Control Bucket Lifecycle Configuration (%s): %s", bucket, err)
	}

	d.SetId(bucket)

	return resourceBucketLifecycleConfigurationRead(ctx, d, meta)
}

func resourceBucketLifecycleConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn(ctx)

	parsedArn, err := arn.Parse(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	if parsedArn.AccountID == "" {
		return diag.Errorf("parsing S3 Control Bucket ARN (%s): unknown format", d.Id())
	}

	output, err := FindBucketLifecycleConfigurationByTwoPartKey(ctx, conn, parsedArn.AccountID, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Control Bucket Lifecycle Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading S3 Control Bucket Lifecycle Configuration (%s): %s", d.Id(), err)
	}

	d.Set("bucket", d.Id())

	if err := d.Set("rule", flattenLifecycleRules(ctx, output.Rules)); err != nil {
		return diag.Errorf("setting rule: %s", err)
	}

	return nil
}

func resourceBucketLifecycleConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn(ctx)

	parsedArn, err := arn.Parse(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	if parsedArn.AccountID == "" {
		return diag.Errorf("parsing S3 Control Bucket ARN (%s): unknown format", d.Id())
	}

	input := &s3control.PutBucketLifecycleConfigurationInput{
		AccountId: aws.String(parsedArn.AccountID),
		Bucket:    aws.String(d.Id()),
		LifecycleConfiguration: &s3control.LifecycleConfiguration{
			Rules: expandLifecycleRules(ctx, d.Get("rule").(*schema.Set).List()),
		},
	}

	_, err = conn.PutBucketLifecycleConfigurationWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("updating S3 Control Bucket Lifecycle Configuration (%s): %s", d.Id(), err)
	}

	return resourceBucketLifecycleConfigurationRead(ctx, d, meta)
}

func resourceBucketLifecycleConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn(ctx)

	parsedArn, err := arn.Parse(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	if parsedArn.AccountID == "" {
		return diag.Errorf("parsing S3 Control Bucket ARN (%s): unknown format", d.Id())
	}

	log.Printf("[DEBUG] Deleting S3 Control Bucket Lifecycle Configuration: %s", d.Id())
	_, err = conn.DeleteBucketLifecycleConfigurationWithContext(ctx, &s3control.DeleteBucketLifecycleConfigurationInput{
		AccountId: aws.String(parsedArn.AccountID),
		Bucket:    aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchLifecycleConfiguration, errCodeNoSuchOutpost) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting S3 Control Bucket Lifecycle Configuration (%s): %s", d.Id(), err)
	}

	return nil
}

func FindBucketLifecycleConfigurationByTwoPartKey(ctx context.Context, conn *s3control.S3Control, accountID, bucket string) (*s3control.GetBucketLifecycleConfigurationOutput, error) {
	input := &s3control.GetBucketLifecycleConfigurationInput{
		AccountId: aws.String(accountID),
		Bucket:    aws.String(bucket),
	}

	output, err := conn.GetBucketLifecycleConfigurationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchLifecycleConfiguration, errCodeNoSuchOutpost) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
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

func expandLifecycleRules(ctx context.Context, tfList []interface{}) []*s3control.LifecycleRule {
	var apiObjects []*s3control.LifecycleRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandLifecycleRule(ctx, tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandLifecycleRule(ctx context.Context, tfMap map[string]interface{}) *s3control.LifecycleRule {
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
		apiObject.Filter = expandLifecycleRuleFilter(ctx, v)
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

func expandLifecycleRuleFilter(ctx context.Context, tfList []interface{}) *s3control.LifecycleRuleFilter {
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
			apiObject.Tag = Tags(tftags.New(ctx, v))[0]
		} else {
			apiObject.And = &s3control.LifecycleRuleAndOperator{
				Prefix: apiObject.Prefix,
				Tags:   Tags(tftags.New(ctx, v)),
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

func flattenLifecycleRules(ctx context.Context, apiObjects []*s3control.LifecycleRule) []interface{} {
	var tfMaps []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfMaps = append(tfMaps, flattenLifecycleRule(ctx, apiObject))
	}

	return tfMaps
}

func flattenLifecycleRule(ctx context.Context, apiObject *s3control.LifecycleRule) map[string]interface{} {
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
		tfMap["filter"] = flattenLifecycleRuleFilter(ctx, v)
	}

	if v := apiObject.ID; v != nil {
		tfMap["id"] = aws.StringValue(v)
	}

	if v := apiObject.Status; v != nil {
		tfMap["status"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenLifecycleRuleFilter(ctx context.Context, apiObject *s3control.LifecycleRuleFilter) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.And != nil {
		if v := apiObject.And.Prefix; v != nil {
			tfMap["prefix"] = aws.StringValue(v)
		}

		if v := apiObject.And.Tags; v != nil {
			tfMap["tags"] = KeyValueTags(ctx, v).IgnoreAWS().Map()
		}
	} else {
		if v := apiObject.Prefix; v != nil {
			tfMap["prefix"] = aws.StringValue(v)
		}

		if v := apiObject.Tag; v != nil {
			tfMap["tags"] = KeyValueTags(ctx, []*s3control.S3Tag{v}).IgnoreAWS().Map()
		}
	}

	return []interface{}{tfMap}
}
