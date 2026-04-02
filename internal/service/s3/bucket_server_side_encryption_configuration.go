// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package s3

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3_bucket_server_side_encryption_configuration", name="Bucket Server Side Encryption Configuration")
// @IdentityAttribute("bucket")
// @IdentityVersion(1)
// @Testing(preIdentityVersion="v6.9.0")
// @Testing(identityVersion="0;v6.10.0")
// @Testing(identityVersion="1;v6.31.0")
// @Testing(checkDestroyNoop=true)
// @Testing(importIgnore="rule.0.bucket_key_enabled")
// @Testing(plannableImportAction="NoOp")
func resourceBucketServerSideEncryptionConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketServerSideEncryptionConfigurationCreate,
		ReadWithoutTimeout:   resourceBucketServerSideEncryptionConfigurationRead,
		UpdateWithoutTimeout: resourceBucketServerSideEncryptionConfigurationUpdate,
		DeleteWithoutTimeout: resourceBucketServerSideEncryptionConfigurationDelete,

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
				ValidateFunc: verify.ValidAccountID,
				Deprecated:   "expected_bucket_owner is deprecated. It will be removed in a future verion of the provider.",
			},
			names.AttrRule: {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"apply_server_side_encryption_by_default": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"kms_master_key_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"sse_algorithm": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ServerSideEncryption](),
									},
								},
							},
						},
						"blocked_encryption_types": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: enum.Validate[awstypes.EncryptionType](),
							},
						},
						"bucket_key_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
		},

		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
			if d.Id() == "" {
				return nil
			}
			if d.HasChange(names.AttrExpectedBucketOwner) {
				o, n := d.GetChange(names.AttrExpectedBucketOwner)
				os, ns := o.(string), n.(string)
				if os == ns {
					return nil
				}
				if os == "" && ns == meta.(*conns.AWSClient).AccountID(ctx) {
					return nil
				}
				return d.ForceNew(names.AttrExpectedBucketOwner)
			}
			return nil
		},
	}
}

func resourceBucketServerSideEncryptionConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Get(names.AttrBucket).(string)
	if isDirectoryBucket(bucket) {
		conn = meta.(*conns.AWSClient).S3ExpressClient(ctx)
	}
	expectedBucketOwner := d.Get(names.AttrExpectedBucketOwner).(string)
	input := s3.PutBucketEncryptionInput{
		Bucket: aws.String(bucket),
		ServerSideEncryptionConfiguration: &awstypes.ServerSideEncryptionConfiguration{
			Rules: expandServerSideEncryptionRules(d.Get(names.AttrRule).(*schema.Set).List()),
		},
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketPropagationTimeout, func(ctx context.Context) (any, error) {
		return conn.PutBucketEncryption(ctx, &input)
	}, errCodeNoSuchBucket, errCodeOperationAborted)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Bucket (%s) Server-side Encryption Configuration: %s", bucket, err)
	}

	d.SetId(createResourceID(bucket, expectedBucketOwner))

	_, err = tfresource.RetryWhenNotFound(ctx, bucketPropagationTimeout, func(ctx context.Context) (any, error) {
		return findServerSideEncryptionConfiguration(ctx, conn, bucket, expectedBucketOwner)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Server-side Encryption Configuration (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceBucketServerSideEncryptionConfigurationRead(ctx, d, meta)...)
}

func resourceBucketServerSideEncryptionConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := parseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if isDirectoryBucket(bucket) {
		conn = meta.(*conns.AWSClient).S3ExpressClient(ctx)
	}

	sse, err := findServerSideEncryptionConfiguration(ctx, conn, bucket, expectedBucketOwner)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] S3 Bucket Server-side Encryption Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket Server-side Encryption Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrBucket, bucket)
	d.Set(names.AttrExpectedBucketOwner, expectedBucketOwner)
	if err := d.Set(names.AttrRule, flattenServerSideEncryptionRules(sse.Rules)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func resourceBucketServerSideEncryptionConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := parseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if isDirectoryBucket(bucket) {
		conn = meta.(*conns.AWSClient).S3ExpressClient(ctx)
	}

	input := s3.PutBucketEncryptionInput{
		Bucket: aws.String(bucket),
		ServerSideEncryptionConfiguration: &awstypes.ServerSideEncryptionConfiguration{
			Rules: expandServerSideEncryptionRules(d.Get(names.AttrRule).(*schema.Set).List()),
		},
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketPropagationTimeout, func(ctx context.Context) (any, error) {
		return conn.PutBucketEncryption(ctx, &input)
	}, errCodeNoSuchBucket, errCodeOperationAborted)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Bucket Server-side Encryption Configuration (%s): %s", d.Id(), err)
	}

	return append(diags, resourceBucketServerSideEncryptionConfigurationRead(ctx, d, meta)...)
}

func resourceBucketServerSideEncryptionConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := parseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if isDirectoryBucket(bucket) {
		conn = meta.(*conns.AWSClient).S3ExpressClient(ctx)
	}

	input := s3.DeleteBucketEncryptionInput{
		Bucket: aws.String(bucket),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.DeleteBucketEncryption(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeServerSideEncryptionConfigurationNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket Server-side Encryption Configuration (%s): %s", d.Id(), err)
	}

	// Don't wait for the SSE configuration to disappear as the bucket now always has one.

	return diags
}

func findServerSideEncryptionConfiguration(ctx context.Context, conn *s3.Client, bucketName, expectedBucketOwner string) (*awstypes.ServerSideEncryptionConfiguration, error) {
	input := s3.GetBucketEncryptionInput{
		Bucket: aws.String(bucketName),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	output, err := conn.GetBucketEncryption(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeServerSideEncryptionConfigurationNotFound) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ServerSideEncryptionConfiguration == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.ServerSideEncryptionConfiguration, nil
}

func expandServerSideEncryptionByDefault(l []any) *awstypes.ServerSideEncryptionByDefault {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]any)
	if !ok {
		return nil
	}

	sse := &awstypes.ServerSideEncryptionByDefault{}

	if v, ok := tfMap["kms_master_key_id"].(string); ok && v != "" {
		sse.KMSMasterKeyID = aws.String(v)
	}

	if v, ok := tfMap["sse_algorithm"].(string); ok && v != "" {
		sse.SSEAlgorithm = awstypes.ServerSideEncryption(v)
	}

	return sse
}

func expandServerSideEncryptionRules(l []any) []awstypes.ServerSideEncryptionRule {
	var rules []awstypes.ServerSideEncryptionRule

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		rule := awstypes.ServerSideEncryptionRule{}

		if v, ok := tfMap["apply_server_side_encryption_by_default"].([]any); ok && len(v) > 0 && v[0] != nil {
			rule.ApplyServerSideEncryptionByDefault = expandServerSideEncryptionByDefault(v)
		}

		if v, ok := tfMap["blocked_encryption_types"].([]any); ok && len(v) > 0 {
			rule.BlockedEncryptionTypes = expandBlockedEncryptionTypes(v)
		}

		if v, ok := tfMap["bucket_key_enabled"].(bool); ok {
			rule.BucketKeyEnabled = aws.Bool(v)
		}

		rules = append(rules, rule)
	}

	return rules
}

func flattenServerSideEncryptionRules(rules []awstypes.ServerSideEncryptionRule) []any {
	var results []any

	for _, rule := range rules {
		m := make(map[string]any)

		if rule.ApplyServerSideEncryptionByDefault != nil {
			m["apply_server_side_encryption_by_default"] = flattenServerSideEncryptionByDefault(rule.ApplyServerSideEncryptionByDefault)
		}

		if rule.BlockedEncryptionTypes != nil {
			if flattened := flattenBlockedEncryptionTypes(rule.BlockedEncryptionTypes); flattened != nil {
				m["blocked_encryption_types"] = flattened
			}
		}

		if rule.BucketKeyEnabled != nil {
			m["bucket_key_enabled"] = aws.ToBool(rule.BucketKeyEnabled)
		}

		results = append(results, m)
	}

	return results
}

func flattenServerSideEncryptionByDefault(sse *awstypes.ServerSideEncryptionByDefault) []any {
	if sse == nil {
		return nil
	}

	m := map[string]any{
		"sse_algorithm": sse.SSEAlgorithm,
	}

	if sse.KMSMasterKeyID != nil {
		m["kms_master_key_id"] = aws.ToString(sse.KMSMasterKeyID)
	}

	return []any{m}
}

func expandBlockedEncryptionTypes(l []any) *awstypes.BlockedEncryptionTypes {
	if len(l) == 0 {
		return nil
	}

	var encryptionTypes []awstypes.EncryptionType
	for _, v := range l {
		encryptionTypes = append(encryptionTypes, awstypes.EncryptionType(v.(string)))
	}

	return &awstypes.BlockedEncryptionTypes{
		EncryptionType: encryptionTypes,
	}
}

func flattenBlockedEncryptionTypes(bet *awstypes.BlockedEncryptionTypes) []any {
	if bet == nil || len(bet.EncryptionType) == 0 {
		return nil
	}

	var result []any
	for _, et := range bet.EncryptionType {
		result = append(result, string(et))
	}

	return result
}

func resourceBucketServerSideEncryptionConfigurationFlatten(_ context.Context, sse *awstypes.ServerSideEncryptionConfiguration, d *schema.ResourceData) error {
	if err := d.Set(names.AttrRule, flattenServerSideEncryptionRules(sse.Rules)); err != nil {
		return fmt.Errorf("setting rule: %w", err)
	}
	return nil
}
