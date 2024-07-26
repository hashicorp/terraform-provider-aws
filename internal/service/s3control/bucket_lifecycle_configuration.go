// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	"github.com/aws/aws-sdk-go-v2/service/s3control/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
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
			names.AttrBucket: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrRule: {
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
						names.AttrFilter: {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrPrefix: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrTags: tftags.TagsSchema(),
								},
							},
						},
						names.AttrID: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Optional: true,
							Default:  types.ExpirationStatusEnabled,
						},
					},
				},
			},
		},
	}
}

func resourceBucketLifecycleConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	bucket := d.Get(names.AttrBucket).(string)
	parsedArn, err := arn.Parse(bucket)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if parsedArn.AccountID == "" {
		return sdkdiag.AppendErrorf(diags, "parsing S3 Control Bucket ARN (%s): unknown format", d.Id())
	}

	input := &s3control.PutBucketLifecycleConfigurationInput{
		AccountId: aws.String(parsedArn.AccountID),
		Bucket:    aws.String(bucket),
		LifecycleConfiguration: &types.LifecycleConfiguration{
			Rules: expandLifecycleRules(ctx, d.Get(names.AttrRule).(*schema.Set).List()),
		},
	}

	_, err = conn.PutBucketLifecycleConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Control Bucket Lifecycle Configuration (%s): %s", bucket, err)
	}

	d.SetId(bucket)

	return append(diags, resourceBucketLifecycleConfigurationRead(ctx, d, meta)...)
}

func resourceBucketLifecycleConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	parsedArn, err := arn.Parse(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if parsedArn.AccountID == "" {
		return sdkdiag.AppendErrorf(diags, "parsing S3 Control Bucket ARN (%s): unknown format", d.Id())
	}

	output, err := findBucketLifecycleConfigurationByTwoPartKey(ctx, conn, parsedArn.AccountID, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Control Bucket Lifecycle Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Control Bucket Lifecycle Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrBucket, d.Id())

	if err := d.Set(names.AttrRule, flattenLifecycleRules(ctx, output.Rules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rule: %s", err)
	}

	return diags
}

func resourceBucketLifecycleConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	parsedArn, err := arn.Parse(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if parsedArn.AccountID == "" {
		return sdkdiag.AppendErrorf(diags, "parsing S3 Control Bucket ARN (%s): unknown format", d.Id())
	}

	input := &s3control.PutBucketLifecycleConfigurationInput{
		AccountId: aws.String(parsedArn.AccountID),
		Bucket:    aws.String(d.Id()),
		LifecycleConfiguration: &types.LifecycleConfiguration{
			Rules: expandLifecycleRules(ctx, d.Get(names.AttrRule).(*schema.Set).List()),
		},
	}

	_, err = conn.PutBucketLifecycleConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Control Bucket Lifecycle Configuration (%s): %s", d.Id(), err)
	}

	return append(diags, resourceBucketLifecycleConfigurationRead(ctx, d, meta)...)
}

func resourceBucketLifecycleConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	parsedArn, err := arn.Parse(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if parsedArn.AccountID == "" {
		return sdkdiag.AppendErrorf(diags, "parsing S3 Control Bucket ARN (%s): unknown format", d.Id())
	}

	log.Printf("[DEBUG] Deleting S3 Control Bucket Lifecycle Configuration: %s", d.Id())
	_, err = conn.DeleteBucketLifecycleConfiguration(ctx, &s3control.DeleteBucketLifecycleConfigurationInput{
		AccountId: aws.String(parsedArn.AccountID),
		Bucket:    aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchLifecycleConfiguration, errCodeNoSuchOutpost) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Control Bucket Lifecycle Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func findBucketLifecycleConfigurationByTwoPartKey(ctx context.Context, conn *s3control.Client, accountID, bucket string) (*s3control.GetBucketLifecycleConfigurationOutput, error) {
	input := &s3control.GetBucketLifecycleConfigurationInput{
		AccountId: aws.String(accountID),
		Bucket:    aws.String(bucket),
	}

	output, err := conn.GetBucketLifecycleConfiguration(ctx, input)

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

func expandAbortIncompleteMultipartUpload(tfList []interface{}) *types.AbortIncompleteMultipartUpload {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &types.AbortIncompleteMultipartUpload{}

	if v, ok := tfMap["days_after_initiation"].(int); ok && v != 0 {
		apiObject.DaysAfterInitiation = int32(v)
	}

	return apiObject
}

func expandLifecycleExpiration(tfList []interface{}) *types.LifecycleExpiration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &types.LifecycleExpiration{}

	if v, ok := tfMap["date"].(string); ok && v != "" {
		parsedDate, err := time.Parse("2006-01-02", v)

		if err == nil {
			apiObject.Date = aws.Time(parsedDate)
		}
	}

	if v, ok := tfMap["days"].(int); ok && v != 0 {
		apiObject.Days = int32(v)
	}

	if v, ok := tfMap["expired_object_delete_marker"].(bool); ok && v {
		apiObject.ExpiredObjectDeleteMarker = v
	}

	return apiObject
}

func expandLifecycleRules(ctx context.Context, tfList []interface{}) []types.LifecycleRule {
	var apiObjects []types.LifecycleRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandLifecycleRule(ctx, tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandLifecycleRule(ctx context.Context, tfMap map[string]interface{}) *types.LifecycleRule {
	if len(tfMap) == 0 {
		return nil
	}

	apiObject := &types.LifecycleRule{}

	if v, ok := tfMap["abort_incomplete_multipart_upload"].([]interface{}); ok && len(v) > 0 {
		apiObject.AbortIncompleteMultipartUpload = expandAbortIncompleteMultipartUpload(v)
	}

	if v, ok := tfMap["expiration"].([]interface{}); ok && len(v) > 0 {
		apiObject.Expiration = expandLifecycleExpiration(v)
	}

	if v, ok := tfMap[names.AttrFilter].([]interface{}); ok && len(v) > 0 {
		apiObject.Filter = expandLifecycleRuleFilter(ctx, v)
	}

	if v, ok := tfMap[names.AttrID].(string); ok && v != "" {
		apiObject.ID = aws.String(v)
	}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		apiObject.Status = types.ExpirationStatus(v)
	}

	// Terraform Plugin SDK sometimes sends map with only empty configuration blocks:
	//   map[abort_incomplete_multipart_upload:[] expiration:[] filter:[] id: status:]
	// This is to prevent this error: InvalidParameter: 1 validation error(s) found.
	//  - missing required field, PutBucketLifecycleConfigurationInput.LifecycleConfiguration.Rules[0].Status.
	if apiObject.ID == nil && apiObject.Status == "" {
		return nil
	}

	return apiObject
}

func expandLifecycleRuleFilter(ctx context.Context, tfList []interface{}) *types.LifecycleRuleFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &types.LifecycleRuleFilter{}

	if v, ok := tfMap[names.AttrPrefix].(string); ok && v != "" {
		apiObject.Prefix = aws.String(v)
	}

	if v, ok := tfMap[names.AttrTags].(map[string]interface{}); ok && len(v) > 0 {
		// See also aws_s3_bucket ReplicationRule.Filter handling
		if len(v) == 1 {
			apiObject.Tag = &tagsS3(tftags.New(ctx, v))[0]
		} else {
			apiObject.And = &types.LifecycleRuleAndOperator{
				Prefix: apiObject.Prefix,
				Tags:   tagsS3(tftags.New(ctx, v)),
			}
			apiObject.Prefix = nil
		}
	}

	return apiObject
}

func flattenAbortIncompleteMultipartUpload(apiObject *types.AbortIncompleteMultipartUpload) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"days_after_initiation": apiObject.DaysAfterInitiation,
	}

	return []interface{}{tfMap}
}

func flattenLifecycleExpiration(apiObject *types.LifecycleExpiration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"days":                         apiObject.Days,
		"expired_object_delete_marker": apiObject.ExpiredObjectDeleteMarker,
	}

	if v := apiObject.Date; v != nil {
		tfMap["date"] = aws.ToTime(v).Format("2006-01-02")
	}

	return []interface{}{tfMap}
}

func flattenLifecycleRules(ctx context.Context, apiObjects []types.LifecycleRule) []interface{} {
	var tfMaps []interface{}

	for _, apiObject := range apiObjects {
		tfMaps = append(tfMaps, flattenLifecycleRule(ctx, apiObject))
	}

	return tfMaps
}

func flattenLifecycleRule(ctx context.Context, apiObject types.LifecycleRule) map[string]interface{} {
	tfMap := map[string]interface{}{
		names.AttrStatus: apiObject.Status,
	}

	if v := apiObject.AbortIncompleteMultipartUpload; v != nil {
		tfMap["abort_incomplete_multipart_upload"] = flattenAbortIncompleteMultipartUpload(v)
	}

	if v := apiObject.Expiration; v != nil {
		tfMap["expiration"] = flattenLifecycleExpiration(v)
	}

	if v := apiObject.Filter; v != nil {
		tfMap[names.AttrFilter] = flattenLifecycleRuleFilter(ctx, v)
	}

	if v := apiObject.ID; v != nil {
		tfMap[names.AttrID] = aws.ToString(v)
	}

	return tfMap
}

func flattenLifecycleRuleFilter(ctx context.Context, apiObject *types.LifecycleRuleFilter) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.And != nil {
		if v := apiObject.And.Prefix; v != nil {
			tfMap[names.AttrPrefix] = aws.ToString(v)
		}

		if v := apiObject.And.Tags; v != nil {
			tfMap[names.AttrTags] = keyValueTagsS3(ctx, v).IgnoreAWS().Map()
		}
	} else {
		if v := apiObject.Prefix; v != nil {
			tfMap[names.AttrPrefix] = aws.ToString(v)
		}

		if v := apiObject.Tag; v != nil {
			tfMap[names.AttrTags] = keyValueTagsS3(ctx, []types.S3Tag{*v}).IgnoreAWS().Map()
		}
	}

	return []interface{}{tfMap}
}
