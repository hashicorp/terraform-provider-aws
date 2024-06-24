// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"log"

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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3_bucket_logging", name="Bucket Logging")
func resourceBucketLogging() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketLoggingCreate,
		ReadWithoutTimeout:   resourceBucketLoggingRead,
		UpdateWithoutTimeout: resourceBucketLoggingUpdate,
		DeleteWithoutTimeout: resourceBucketLoggingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
			"target_bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"target_grant": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"grantee": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDisplayName: {
										Type:     schema.TypeString,
										Computed: true,
									},
									"email_address": {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrID: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrType: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.Type](),
									},
									names.AttrURI: {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"permission": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.BucketLogsPermission](),
						},
					},
				},
			},
			"target_object_key_format": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"partitioned_prefix": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"partition_date_source": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.PartitionDateSource](),
									},
								},
							},
							ExactlyOneOf: []string{"target_object_key_format.0.partitioned_prefix", "target_object_key_format.0.simple_prefix"},
						},
						"simple_prefix": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{},
							},
							ExactlyOneOf: []string{"target_object_key_format.0.partitioned_prefix", "target_object_key_format.0.simple_prefix"},
						},
					},
				},
			},
			"target_prefix": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceBucketLoggingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Get(names.AttrBucket).(string)
	expectedBucketOwner := d.Get(names.AttrExpectedBucketOwner).(string)
	input := &s3.PutBucketLoggingInput{
		Bucket: aws.String(bucket),
		BucketLoggingStatus: &types.BucketLoggingStatus{
			LoggingEnabled: &types.LoggingEnabled{
				TargetBucket: aws.String(d.Get("target_bucket").(string)),
				TargetPrefix: aws.String(d.Get("target_prefix").(string)),
			},
		},
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	if v, ok := d.GetOk("target_grant"); ok && v.(*schema.Set).Len() > 0 {
		input.BucketLoggingStatus.LoggingEnabled.TargetGrants = expandTargetGrants(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("target_object_key_format"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.BucketLoggingStatus.LoggingEnabled.TargetObjectKeyFormat = expandTargetObjectKeyFormat(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return conn.PutBucketLogging(ctx, input)
	}, errCodeNoSuchBucket)

	if tfawserr.ErrMessageContains(err, errCodeInvalidArgument, "BucketLoggingStatus is not valid, expected CreateBucketConfiguration") {
		err = errDirectoryBucket(err)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Bucket (%s) Logging: %s", bucket, err)
	}

	d.SetId(CreateResourceID(bucket, expectedBucketOwner))

	_, err = tfresource.RetryWhenNotFound(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return findLoggingEnabled(ctx, conn, bucket, expectedBucketOwner)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Logging (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceBucketLoggingRead(ctx, d, meta)...)
}

func resourceBucketLoggingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	loggingEnabled, err := findLoggingEnabled(ctx, conn, bucket, expectedBucketOwner)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket Logging (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket Logging (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrBucket, bucket)
	d.Set(names.AttrExpectedBucketOwner, expectedBucketOwner)
	d.Set("target_bucket", loggingEnabled.TargetBucket)
	if err := d.Set("target_grant", flattenTargetGrants(loggingEnabled.TargetGrants)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting target_grant: %s", err)
	}
	if loggingEnabled.TargetObjectKeyFormat != nil {
		if err := d.Set("target_object_key_format", []interface{}{flattenTargetObjectKeyFormat(loggingEnabled.TargetObjectKeyFormat)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting target_object_key_format: %s", err)
		}
	} else {
		d.Set("target_object_key_format", nil)
	}
	d.Set("target_prefix", loggingEnabled.TargetPrefix)

	return diags
}

func resourceBucketLoggingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3.PutBucketLoggingInput{
		Bucket: aws.String(bucket),
		BucketLoggingStatus: &types.BucketLoggingStatus{
			LoggingEnabled: &types.LoggingEnabled{
				TargetBucket: aws.String(d.Get("target_bucket").(string)),
				TargetPrefix: aws.String(d.Get("target_prefix").(string)),
			},
		},
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	if v, ok := d.GetOk("target_grant"); ok && v.(*schema.Set).Len() > 0 {
		input.BucketLoggingStatus.LoggingEnabled.TargetGrants = expandTargetGrants(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("target_object_key_format"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.BucketLoggingStatus.LoggingEnabled.TargetObjectKeyFormat = expandTargetObjectKeyFormat(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err = conn.PutBucketLogging(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Bucket Logging (%s): %s", d.Id(), err)
	}

	return append(diags, resourceBucketLoggingRead(ctx, d, meta)...)
}

func resourceBucketLoggingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3.PutBucketLoggingInput{
		Bucket:              aws.String(bucket),
		BucketLoggingStatus: &types.BucketLoggingStatus{},
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.PutBucketLogging(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket Logging (%s): %s", d.Id(), err)
	}

	// Don't wait for the logging to disappear as it still exists after update.

	return diags
}

func findLoggingEnabled(ctx context.Context, conn *s3.Client, bucketName, expectedBucketOwner string) (*types.LoggingEnabled, error) {
	input := &s3.GetBucketLoggingInput{
		Bucket: aws.String(bucketName),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	output, err := conn.GetBucketLogging(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.LoggingEnabled == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.LoggingEnabled, nil
}

func expandTargetGrants(l []interface{}) []types.TargetGrant {
	var grants []types.TargetGrant

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		grant := types.TargetGrant{}

		if v, ok := tfMap["grantee"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			grant.Grantee = expandLoggingGrantee(v)
		}

		if v, ok := tfMap["permission"].(string); ok && v != "" {
			grant.Permission = types.BucketLogsPermission(v)
		}

		grants = append(grants, grant)
	}

	return grants
}

func expandLoggingGrantee(l []interface{}) *types.Grantee {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	grantee := &types.Grantee{}

	if v, ok := tfMap[names.AttrDisplayName].(string); ok && v != "" {
		grantee.DisplayName = aws.String(v)
	}

	if v, ok := tfMap["email_address"].(string); ok && v != "" {
		grantee.EmailAddress = aws.String(v)
	}

	if v, ok := tfMap[names.AttrID].(string); ok && v != "" {
		grantee.ID = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		grantee.Type = types.Type(v)
	}

	if v, ok := tfMap[names.AttrURI].(string); ok && v != "" {
		grantee.URI = aws.String(v)
	}

	return grantee
}

func flattenTargetGrants(grants []types.TargetGrant) []interface{} {
	var results []interface{}

	for _, grant := range grants {
		m := map[string]interface{}{
			"permission": grant.Permission,
		}

		if grant.Grantee != nil {
			m["grantee"] = flattenLoggingGrantee(grant.Grantee)
		}

		results = append(results, m)
	}

	return results
}

func flattenLoggingGrantee(g *types.Grantee) []interface{} {
	if g == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrType: g.Type,
	}

	if g.DisplayName != nil {
		m[names.AttrDisplayName] = aws.ToString(g.DisplayName)
	}

	if g.EmailAddress != nil {
		m["email_address"] = aws.ToString(g.EmailAddress)
	}

	if g.ID != nil {
		m[names.AttrID] = aws.ToString(g.ID)
	}

	if g.URI != nil {
		m[names.AttrURI] = aws.ToString(g.URI)
	}

	return []interface{}{m}
}

func expandTargetObjectKeyFormat(tfMap map[string]interface{}) *types.TargetObjectKeyFormat {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.TargetObjectKeyFormat{}

	if v, ok := tfMap["partitioned_prefix"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.PartitionedPrefix = expandPartitionedPrefix(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["simple_prefix"]; ok && len(v.([]interface{})) > 0 {
		apiObject.SimplePrefix = &types.SimplePrefix{}
	}

	return apiObject
}

func expandPartitionedPrefix(tfMap map[string]interface{}) *types.PartitionedPrefix {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PartitionedPrefix{}

	if v, ok := tfMap["partition_date_source"].(string); ok && v != "" {
		apiObject.PartitionDateSource = types.PartitionDateSource(v)
	}

	return apiObject
}

func flattenTargetObjectKeyFormat(apiObject *types.TargetObjectKeyFormat) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.PartitionedPrefix; v != nil {
		tfMap["partitioned_prefix"] = []interface{}{flattenPartitionedPrefix(v)}
	}

	if apiObject.SimplePrefix != nil {
		tfMap["simple_prefix"] = make([]map[string]interface{}, 1)
	}

	return tfMap
}

func flattenPartitionedPrefix(apiObject *types.PartitionedPrefix) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"partition_date_source": apiObject.PartitionDateSource,
	}

	return tfMap
}
