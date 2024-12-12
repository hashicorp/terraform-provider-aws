// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3_bucket_lifecycle_configuration", name="Bucket Lifecycle Configuration")
func resourceBucketLifecycleConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketLifecycleConfigurationCreate,
		ReadWithoutTimeout:   resourceBucketLifecycleConfigurationRead,
		UpdateWithoutTimeout: resourceBucketLifecycleConfigurationUpdate,
		DeleteWithoutTimeout: resourceBucketLifecycleConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrBucket: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 63),
			},
			names.AttrExpectedBucketOwner: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			names.AttrRule: {
				Type:     schema.TypeList,
				Required: true,
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
										Optional: true,
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
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidUTCTimestamp,
									},
									"days": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  0, // API returns 0
									},
									"expired_object_delete_marker": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true, // API returns false; conflicts with date and days
									},
								},
							},
						},
						names.AttrFilter: {
							Type:     schema.TypeList,
							Optional: true,
							// If neither the filter block nor the prefix parameter in the rule are specified,
							// we apply the Default behavior from v3.x of the provider (Filter with empty string Prefix),
							// which will thus return a Filter in the GetBucketLifecycleConfiguration request and
							// require diff suppression.
							DiffSuppressFunc: suppressMissingFilterConfigurationBlock,
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"and": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object_size_greater_than": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(0),
												},
												"object_size_less_than": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
												names.AttrPrefix: {
													Type:     schema.TypeString,
													Optional: true,
												},
												names.AttrTags: tftags.TagsSchema(),
											},
										},
									},
									"object_size_greater_than": {
										Type:     nullable.TypeNullableInt,
										Optional: true,
									},
									"object_size_less_than": {
										Type:     nullable.TypeNullableInt,
										Optional: true,
									},
									names.AttrPrefix: {
										Type:     schema.TypeString,
										Optional: true,
									},
									"tag": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrKey: {
													Type:     schema.TypeString,
													Required: true,
												},
												names.AttrValue: {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						names.AttrID: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"noncurrent_version_expiration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"newer_noncurrent_versions": {
										Type:         nullable.TypeNullableInt,
										Optional:     true,
										ValidateFunc: nullable.ValidateTypeStringNullableIntAtLeast(1),
									},
									"noncurrent_days": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"noncurrent_version_transition": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"newer_noncurrent_versions": {
										Type:         nullable.TypeNullableInt,
										Optional:     true,
										ValidateFunc: nullable.ValidateTypeStringNullableIntAtLeast(1),
									},
									"noncurrent_days": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
									},
									names.AttrStorageClass: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.TransitionStorageClass](),
									},
								},
							},
						},
						names.AttrPrefix: {
							Type:       schema.TypeString,
							Optional:   true,
							Deprecated: "Use filter instead",
						},
						names.AttrStatus: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(lifecycleRuleStatus_Values(), false),
						},
						"transition": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"date": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidUTCTimestamp,
									},
									"days": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
									},
									names.AttrStorageClass: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.TransitionStorageClass](),
									},
								},
							},
						},
					},
				},
			},
			"transition_default_minimum_object_size": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.TransitionDefaultMinimumObjectSize](),
			},
		},
	}
}

func resourceBucketLifecycleConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Get(names.AttrBucket).(string)
	if isDirectoryBucket(bucket) {
		conn = meta.(*conns.AWSClient).S3ExpressClient(ctx)
	}
	expectedBucketOwner := d.Get(names.AttrExpectedBucketOwner).(string)
	rules := expandLifecycleRules(ctx, d.Get(names.AttrRule).([]interface{}))
	input := &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucket),
		LifecycleConfiguration: &types.BucketLifecycleConfiguration{
			Rules: rules,
		},
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	if v, ok := d.GetOk("transition_default_minimum_object_size"); ok {
		input.TransitionDefaultMinimumObjectSize = types.TransitionDefaultMinimumObjectSize(v.(string))
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return conn.PutBucketLifecycleConfiguration(ctx, input)
	}, errCodeNoSuchBucket)

	if tfawserr.ErrMessageContains(err, errCodeInvalidArgument, "LifecycleConfiguration is not valid, expected CreateBucketConfiguration") {
		err = errDirectoryBucket(err)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Bucket (%s) Lifecycle Configuration: %s", bucket, err)
	}

	d.SetId(CreateResourceID(bucket, expectedBucketOwner))

	if _, err := waitLifecycleRulesEquals(ctx, conn, bucket, expectedBucketOwner, rules, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Lifecycle Configuration (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceBucketLifecycleConfigurationRead(ctx, d, meta)...)
}

func resourceBucketLifecycleConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if isDirectoryBucket(bucket) {
		conn = meta.(*conns.AWSClient).S3ExpressClient(ctx)
	}

	const (
		lifecycleConfigurationExtraRetryDelay    = 5 * time.Second
		lifecycleConfigurationRulesSteadyTimeout = 2 * time.Minute
	)
	var lastOutput, output *s3.GetBucketLifecycleConfigurationOutput

	err = retry.RetryContext(ctx, lifecycleConfigurationRulesSteadyTimeout, func() *retry.RetryError {
		var err error

		time.Sleep(lifecycleConfigurationExtraRetryDelay)

		output, err = findBucketLifecycleConfiguration(ctx, conn, bucket, expectedBucketOwner)

		if d.IsNewResource() && tfresource.NotFound(err) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		if lastOutput == nil || !lifecycleRulesEqual(lastOutput.Rules, output.Rules) {
			lastOutput = output
			return retry.RetryableError(fmt.Errorf("S3 Bucket Lifecycle Configuration (%s) has not stablized; retrying", d.Id()))
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = findBucketLifecycleConfiguration(ctx, conn, bucket, expectedBucketOwner)
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket Lifecycle Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket Lifecycle Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrBucket, bucket)
	d.Set(names.AttrExpectedBucketOwner, expectedBucketOwner)
	if err := d.Set(names.AttrRule, flattenLifecycleRules(ctx, output.Rules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rule: %s", err)
	}
	d.Set("transition_default_minimum_object_size", output.TransitionDefaultMinimumObjectSize)

	return diags
}

func resourceBucketLifecycleConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if isDirectoryBucket(bucket) {
		conn = meta.(*conns.AWSClient).S3ExpressClient(ctx)
	}

	rules := expandLifecycleRules(ctx, d.Get(names.AttrRule).([]interface{}))
	input := &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucket),
		LifecycleConfiguration: &types.BucketLifecycleConfiguration{
			Rules: rules,
		},
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	if v, ok := d.GetOk("transition_default_minimum_object_size"); ok {
		input.TransitionDefaultMinimumObjectSize = types.TransitionDefaultMinimumObjectSize(v.(string))
	}

	_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return conn.PutBucketLifecycleConfiguration(ctx, input)
	}, errCodeNoSuchLifecycleConfiguration)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Bucket Lifecycle Configuration (%s): %s", d.Id(), err)
	}

	if _, err := waitLifecycleRulesEquals(ctx, conn, bucket, expectedBucketOwner, rules, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Lifecycle Configuration (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceBucketLifecycleConfigurationRead(ctx, d, meta)...)
}

func resourceBucketLifecycleConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if isDirectoryBucket(bucket) {
		conn = meta.(*conns.AWSClient).S3ExpressClient(ctx)
	}

	input := &s3.DeleteBucketLifecycleInput{
		Bucket: aws.String(bucket),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.DeleteBucketLifecycle(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchLifecycleConfiguration) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket Lifecycle Configuration (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return findBucketLifecycleConfiguration(ctx, conn, bucket, expectedBucketOwner)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Lifecyle Configuration (%s) delete: %s", d.Id(), err)
	}

	return diags
}

// suppressMissingFilterConfigurationBlock suppresses the diff that results from an omitted
// filter configuration block and one returned from the S3 API.
// To work around the issue, https://github.com/hashicorp/terraform-plugin-sdk/issues/743,
// this method only looks for changes in the "filter.#" value and not its nested fields
// which are incorrectly suppressed when using the verify.SuppressMissingOptionalConfigurationBlock method.
func suppressMissingFilterConfigurationBlock(k, old, new string, d *schema.ResourceData) bool {
	if strings.HasSuffix(k, "filter.#") {
		oraw, nraw := d.GetChange(k)
		o, n := oraw.(int), nraw.(int)

		if o == 1 && n == 0 {
			return true
		}

		if o == 1 && n == 1 {
			return old == "1" && new == "0"
		}

		return false
	}
	return false
}

func findBucketLifecycleConfiguration(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string) (*s3.GetBucketLifecycleConfigurationOutput, error) {
	input := &s3.GetBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucket),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	output, err := conn.GetBucketLifecycleConfiguration(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchLifecycleConfiguration) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Rules) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func lifecycleRulesEqual(rules1, rules2 []types.LifecycleRule) bool {
	if len(rules1) != len(rules2) {
		return false
	}

	for _, rule1 := range rules1 {
		// We consider 2 LifecycleRules equal if their IDs and Statuses are equal.
		if !slices.ContainsFunc(rules2, func(rule2 types.LifecycleRule) bool {
			return aws.ToString(rule1.ID) == aws.ToString(rule2.ID) && rule1.Status == rule2.Status
		}) {
			return false
		}
	}

	return true
}

func statusLifecycleRulesEquals(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string, rules []types.LifecycleRule) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findBucketLifecycleConfiguration(ctx, conn, bucket, expectedBucketOwner)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(lifecycleRulesEqual(output.Rules, rules)), nil
	}
}

func waitLifecycleRulesEquals(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string, rules []types.LifecycleRule, timeout time.Duration) ([]types.LifecycleRule, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Target:                    []string{strconv.FormatBool(true)},
		Refresh:                   statusLifecycleRulesEquals(ctx, conn, bucket, expectedBucketOwner, rules),
		Timeout:                   timeout,
		MinTimeout:                10 * time.Second,
		ContinuousTargetOccurence: 3,
		NotFoundChecks:            20,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]types.LifecycleRule); ok {
		return output, err
	}

	return nil, err
}

const (
	lifecycleRuleStatusDisabled = "Disabled"
	lifecycleRuleStatusEnabled  = "Enabled"
)

func lifecycleRuleStatus_Values() []string {
	return []string{
		lifecycleRuleStatusDisabled,
		lifecycleRuleStatusEnabled,
	}
}

func expandLifecycleRules(ctx context.Context, tfList []interface{}) []types.LifecycleRule {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	var apiObjects []types.LifecycleRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := types.LifecycleRule{}

		if v, ok := tfMap["abort_incomplete_multipart_upload"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			apiObject.AbortIncompleteMultipartUpload = expandAbortIncompleteMultipartUpload(v[0].(map[string]interface{}))
		}

		if v, ok := tfMap["expiration"].([]interface{}); ok && len(v) > 0 {
			apiObject.Expiration = expandLifecycleExpiration(v)
		}

		if v, ok := tfMap[names.AttrFilter].([]interface{}); ok && len(v) > 0 {
			apiObject.Filter = expandLifecycleRuleFilter(ctx, v)
		}

		if v, ok := tfMap[names.AttrPrefix].(string); ok && apiObject.Filter == nil {
			// If neither the filter block nor the prefix are specified,
			// apply the Default behavior from v3.x of the provider;
			// otherwise, set the prefix as specified in Terraform.
			if v == "" {
				apiObject.Filter = &types.LifecycleRuleFilter{
					Prefix: aws.String(v),
				}
			} else {
				apiObject.Prefix = aws.String(v)
			}
		}

		if v, ok := tfMap[names.AttrID].(string); ok {
			apiObject.ID = aws.String(v)
		}

		if v, ok := tfMap["noncurrent_version_expiration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			apiObject.NoncurrentVersionExpiration = expandNoncurrentVersionExpiration(v[0].(map[string]interface{}))
		}

		if v, ok := tfMap["noncurrent_version_transition"].(*schema.Set); ok && v.Len() > 0 {
			apiObject.NoncurrentVersionTransitions = expandNoncurrentVersionTransitions(v.List())
		}

		if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
			apiObject.Status = types.ExpirationStatus(v)
		}

		if v, ok := tfMap["transition"].(*schema.Set); ok && v.Len() > 0 {
			apiObject.Transitions = expandTransitions(v.List())
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandAbortIncompleteMultipartUpload(tfMap map[string]interface{}) *types.AbortIncompleteMultipartUpload {
	if len(tfMap) == 0 {
		return nil
	}

	result := &types.AbortIncompleteMultipartUpload{}

	if v, ok := tfMap["days_after_initiation"].(int); ok {
		result.DaysAfterInitiation = aws.Int32(int32(v))
	}

	return result
}

func expandLifecycleExpiration(tfList []interface{}) *types.LifecycleExpiration {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &types.LifecycleExpiration{}

	if tfList[0] == nil {
		return apiObject
	}

	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["date"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v)
		apiObject.Date = aws.Time(t)
	}

	if v, ok := tfMap["days"].(int); ok && v > 0 {
		apiObject.Days = aws.Int32(int32(v))
	}

	// This cannot be specified with Days or Date.
	if v, ok := tfMap["expired_object_delete_marker"].(bool); ok && apiObject.Date == nil && aws.ToInt32(apiObject.Days) == 0 {
		apiObject.ExpiredObjectDeleteMarker = aws.Bool(v)
	}

	return apiObject
}

func expandLifecycleRuleFilter(ctx context.Context, tfList []interface{}) *types.LifecycleRuleFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	var apiObject *types.LifecycleRuleFilter

	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["and"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject = &types.LifecycleRuleFilter{
			And: expandLifecycleRuleAndOperator(ctx, v[0].(map[string]interface{})),
		}
	}

	if v, null, _ := nullable.Int(tfMap["object_size_greater_than"].(string)).ValueInt64(); !null && v >= 0 {
		apiObject = &types.LifecycleRuleFilter{
			ObjectSizeGreaterThan: aws.Int64(v),
		}
	}

	if v, null, _ := nullable.Int(tfMap["object_size_less_than"].(string)).ValueInt64(); !null && v > 0 {
		apiObject = &types.LifecycleRuleFilter{
			ObjectSizeLessThan: aws.Int64(v),
		}
	}

	if v, ok := tfMap["tag"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject = &types.LifecycleRuleFilter{
			Tag: expandTag(v[0].(map[string]interface{})),
		}
	}

	// Per AWS S3 API, "A Filter must have exactly one of Prefix, Tag, or And specified";
	// Specifying more than one of the listed parameters results in a MalformedXML error.
	// In practice, this also includes ObjectSizeGreaterThan and ObjectSizeLessThan.
	if v, ok := tfMap[names.AttrPrefix].(string); ok && apiObject == nil {
		apiObject = &types.LifecycleRuleFilter{
			Prefix: aws.String(v),
		}
	}

	return apiObject
}

func expandLifecycleRuleAndOperator(ctx context.Context, tfMap map[string]interface{}) *types.LifecycleRuleAndOperator {
	if len(tfMap) == 0 {
		return nil
	}

	apiObject := &types.LifecycleRuleAndOperator{}

	if v, ok := tfMap["object_size_greater_than"].(int); ok && v > 0 {
		apiObject.ObjectSizeGreaterThan = aws.Int64(int64(v))
	}

	if v, ok := tfMap["object_size_less_than"].(int); ok && v > 0 {
		apiObject.ObjectSizeLessThan = aws.Int64(int64(v))
	}

	if v, ok := tfMap[names.AttrPrefix].(string); ok {
		apiObject.Prefix = aws.String(v)
	}

	if v, ok := tfMap[names.AttrTags].(map[string]interface{}); ok && len(v) > 0 {
		if tags := Tags(tftags.New(ctx, v).IgnoreAWS()); len(tags) > 0 {
			apiObject.Tags = tags
		}
	}

	return apiObject
}

func expandTag(tfMap map[string]interface{}) *types.Tag {
	if len(tfMap) == 0 {
		return nil
	}

	apiObject := &types.Tag{}

	if v, ok := tfMap[names.AttrKey].(string); ok {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func expandNoncurrentVersionExpiration(tfMap map[string]interface{}) *types.NoncurrentVersionExpiration {
	if len(tfMap) == 0 {
		return nil
	}

	apiObject := &types.NoncurrentVersionExpiration{}

	if v, null, _ := nullable.Int(tfMap["newer_noncurrent_versions"].(string)).ValueInt32(); !null && v > 0 {
		apiObject.NewerNoncurrentVersions = aws.Int32(v)
	}

	if v, ok := tfMap["noncurrent_days"].(int); ok {
		apiObject.NoncurrentDays = aws.Int32(int32(v))
	}

	return apiObject
}

func expandNoncurrentVersionTransitions(tfList []interface{}) []types.NoncurrentVersionTransition {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	var apiObjects []types.NoncurrentVersionTransition

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := types.NoncurrentVersionTransition{}

		if v, null, _ := nullable.Int(tfMap["newer_noncurrent_versions"].(string)).ValueInt32(); !null && v > 0 {
			apiObject.NewerNoncurrentVersions = aws.Int32(v)
		}

		if v, ok := tfMap["noncurrent_days"].(int); ok {
			apiObject.NoncurrentDays = aws.Int32(int32(v))
		}

		if v, ok := tfMap[names.AttrStorageClass].(string); ok && v != "" {
			apiObject.StorageClass = types.TransitionStorageClass(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandTransitions(tfList []interface{}) []types.Transition {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	var apiObjects []types.Transition

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := types.Transition{}

		if v, ok := tfMap["date"].(string); ok && v != "" {
			t, _ := time.Parse(time.RFC3339, v)
			apiObject.Date = aws.Time(t)
		}

		// Only one of "date" and "days" can be configured
		// so only set the transition.Days value when transition.Date is nil
		// By default, tfMap["days"] = 0 if not explicitly configured in terraform.
		if v, ok := tfMap["days"].(int); ok && v >= 0 && apiObject.Date == nil {
			apiObject.Days = aws.Int32(int32(v))
		}

		if v, ok := tfMap[names.AttrStorageClass].(string); ok && v != "" {
			apiObject.StorageClass = types.TransitionStorageClass(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenLifecycleRules(ctx context.Context, apiObjects []types.LifecycleRule) []interface{} {
	if len(apiObjects) == 0 {
		return []interface{}{}
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			names.AttrStatus: apiObject.Status,
		}

		if apiObject.AbortIncompleteMultipartUpload != nil {
			tfMap["abort_incomplete_multipart_upload"] = flattenAbortIncompleteMultipartUpload(apiObject.AbortIncompleteMultipartUpload)
		}

		if apiObject.Expiration != nil {
			tfMap["expiration"] = flattenLifecycleExpiration(apiObject.Expiration)
		}

		if apiObject.Filter != nil {
			tfMap[names.AttrFilter] = flattenLifecycleRuleFilter(ctx, apiObject.Filter)
		}

		if apiObject.ID != nil {
			tfMap[names.AttrID] = aws.ToString(apiObject.ID)
		}

		if apiObject.NoncurrentVersionExpiration != nil {
			tfMap["noncurrent_version_expiration"] = flattenNoncurrentVersionExpiration(apiObject.NoncurrentVersionExpiration)
		}

		if apiObject.NoncurrentVersionTransitions != nil {
			tfMap["noncurrent_version_transition"] = flattenNoncurrentVersionTransitions(apiObject.NoncurrentVersionTransitions)
		}

		if apiObject.Prefix != nil {
			tfMap[names.AttrPrefix] = aws.ToString(apiObject.Prefix)
		}

		if apiObject.Transitions != nil {
			tfMap["transition"] = flattenTransitions(apiObject.Transitions)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenAbortIncompleteMultipartUpload(apiObject *types.AbortIncompleteMultipartUpload) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := make(map[string]interface{})

	if apiObject.DaysAfterInitiation != nil {
		tfMap["days_after_initiation"] = aws.ToInt32(apiObject.DaysAfterInitiation)
	}

	return []interface{}{tfMap}
}

func flattenLifecycleExpiration(apiObject *types.LifecycleExpiration) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := make(map[string]interface{})

	if apiObject.Date != nil {
		tfMap["date"] = apiObject.Date.Format(time.RFC3339)
	}

	if apiObject.Days != nil {
		tfMap["days"] = aws.ToInt32(apiObject.Days)
	}

	if apiObject.ExpiredObjectDeleteMarker != nil {
		tfMap["expired_object_delete_marker"] = aws.ToBool(apiObject.ExpiredObjectDeleteMarker)
	}

	return []interface{}{tfMap}
}

func flattenLifecycleRuleFilter(ctx context.Context, apiObject *types.LifecycleRuleFilter) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.And; v != nil {
		tfMap["and"] = flattenLifecycleRuleAndOperator(ctx, v)
	}

	if v := apiObject.ObjectSizeGreaterThan; v != nil {
		tfMap["object_size_greater_than"] = flex.Int64ToStringValue(v)
	}

	if v := apiObject.ObjectSizeLessThan; v != nil {
		tfMap["object_size_less_than"] = flex.Int64ToStringValue(v)
	}

	if v := apiObject.Prefix; v != nil {
		tfMap[names.AttrPrefix] = aws.ToString(v)
	}

	if v := apiObject.Tag; v != nil {
		tfMap["tag"] = flattenTag(v)
	}

	return []interface{}{tfMap}
}

func flattenLifecycleRuleAndOperator(ctx context.Context, apiObject *types.LifecycleRuleAndOperator) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"object_size_greater_than": aws.ToInt64(apiObject.ObjectSizeGreaterThan),
		"object_size_less_than":    aws.ToInt64(apiObject.ObjectSizeLessThan),
	}

	if v := apiObject.Prefix; v != nil {
		tfMap[names.AttrPrefix] = aws.ToString(v)
	}

	if v := apiObject.Tags; v != nil {
		tfMap[names.AttrTags] = keyValueTags(ctx, v).IgnoreAWS().Map()
	}

	return []interface{}{tfMap}
}

func flattenTag(apiObject *types.Tag) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.Key; v != nil {
		tfMap[names.AttrKey] = aws.ToString(v)
	}

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func flattenNoncurrentVersionExpiration(apiObject *types.NoncurrentVersionExpiration) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := make(map[string]interface{})

	if apiObject.NewerNoncurrentVersions != nil {
		tfMap["newer_noncurrent_versions"] = flex.Int32ToStringValue(apiObject.NewerNoncurrentVersions)
	}

	if apiObject.NoncurrentDays != nil {
		tfMap["noncurrent_days"] = aws.ToInt32(apiObject.NoncurrentDays)
	}

	return []interface{}{tfMap}
}

func flattenNoncurrentVersionTransitions(apiObjects []types.NoncurrentVersionTransition) []interface{} {
	if len(apiObjects) == 0 {
		return []interface{}{}
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			names.AttrStorageClass: apiObject.StorageClass,
		}

		if apiObject.NewerNoncurrentVersions != nil {
			tfMap["newer_noncurrent_versions"] = flex.Int32ToStringValue(apiObject.NewerNoncurrentVersions)
		}

		if apiObject.NoncurrentDays != nil {
			tfMap["noncurrent_days"] = aws.ToInt32(apiObject.NoncurrentDays)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenTransitions(apiObjects []types.Transition) []interface{} {
	if len(apiObjects) == 0 {
		return []interface{}{}
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			names.AttrStorageClass: apiObject.StorageClass,
		}

		if apiObject.Date != nil {
			tfMap["date"] = apiObject.Date.Format(time.RFC3339)
		}

		if apiObject.Days != nil {
			tfMap["days"] = aws.ToInt32(apiObject.Days)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
