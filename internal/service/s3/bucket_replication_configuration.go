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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3_bucket_replication_configuration", name="Bucket Replication Configuration")
func resourceBucketReplicationConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketReplicationConfigurationCreate,
		ReadWithoutTimeout:   resourceBucketReplicationConfigurationRead,
		UpdateWithoutTimeout: resourceBucketReplicationConfigurationUpdate,
		DeleteWithoutTimeout: resourceBucketReplicationConfigurationDelete,

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
			names.AttrRole: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrRule: {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1000,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delete_marker_replication": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrStatus: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.DeleteMarkerReplicationStatus](),
									},
								},
							},
						},
						names.AttrDestination: {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"access_control_translation": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrOwner: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[types.OwnerOverride](),
												},
											},
										},
									},
									"account": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidAccountID,
									},
									names.AttrBucket: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									names.AttrEncryptionConfiguration: {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"replica_kms_key_id": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									"metrics": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"event_threshold": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"minutes": {
																Type:     schema.TypeInt,
																Required: true,
																// Currently, the S3 API only supports 15 minutes;
																// however, to account for future changes, validation
																// is left at positive integers.
																ValidateFunc: validation.IntAtLeast(0),
															},
														},
													},
												},
												names.AttrStatus: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[types.MetricsStatus](),
												},
											},
										},
									},
									"replication_time": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrStatus: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[types.ReplicationTimeStatus](),
												},
												"time": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"minutes": {
																Type:     schema.TypeInt,
																Required: true,
																// Currently, the S3 API only supports 15 minutes;
																// however, to account for future changes, validation
																// is left at positive integers.
																ValidateFunc: validation.IntAtLeast(0),
															},
														},
													},
												},
											},
										},
									},
									names.AttrStorageClass: {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.StorageClass](),
									},
								},
							},
						},
						"existing_object_replication": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrStatus: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.ExistingObjectReplicationStatus](),
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
									"and": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrPrefix: {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 1024),
												},
												names.AttrTags: tftags.TagsSchema(),
											},
										},
									},
									names.AttrPrefix: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 1024),
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
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						names.AttrPrefix: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
							Deprecated:   "Use filter instead",
						},
						names.AttrPriority: {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"source_selection_criteria": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"replica_modifications": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrStatus: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[types.ReplicaModificationsStatus](),
												},
											},
										},
									},
									"sse_kms_encrypted_objects": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrStatus: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[types.SseKmsEncryptedObjectsStatus](),
												},
											},
										},
									},
								},
							},
						},
						names.AttrStatus: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.ReplicationRuleStatus](),
						},
					},
				},
			},
			"token": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceBucketReplicationConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Get(names.AttrBucket).(string)
	input := &s3.PutBucketReplicationInput{
		Bucket: aws.String(bucket),
		ReplicationConfiguration: &types.ReplicationConfiguration{
			Role:  aws.String(d.Get(names.AttrRole).(string)),
			Rules: expandReplicationRules(ctx, d.Get(names.AttrRule).([]interface{})),
		},
	}

	if v, ok := d.GetOk("token"); ok {
		input.Token = aws.String(v.(string))
	}

	err := retry.RetryContext(ctx, bucketPropagationTimeout, func() *retry.RetryError {
		_, err := conn.PutBucketReplication(ctx, input)

		if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) || tfawserr.ErrMessageContains(err, errCodeInvalidRequest, "Versioning must be 'Enabled' on the bucket") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutBucketReplication(ctx, input)
	}

	if tfawserr.ErrMessageContains(err, errCodeInvalidArgument, "ReplicationConfiguration is not valid, expected CreateBucketConfiguration") {
		err = errDirectoryBucket(err)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Bucket (%s) Replication Configuration: %s", bucket, err)
	}

	d.SetId(bucket)

	_, err = tfresource.RetryWhenNotFound(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return findReplicationConfiguration(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Replication Configuration (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceBucketReplicationConfigurationRead(ctx, d, meta)...)
}

func resourceBucketReplicationConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	rc, err := findReplicationConfiguration(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket Replication Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket Replication Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrBucket, d.Id())
	d.Set(names.AttrRole, rc.Role)
	if err := d.Set(names.AttrRule, flattenReplicationRules(ctx, rc.Rules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rule: %s", err)
	}

	return diags
}

func resourceBucketReplicationConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	input := &s3.PutBucketReplicationInput{
		Bucket: aws.String(d.Id()),
		ReplicationConfiguration: &types.ReplicationConfiguration{
			Role:  aws.String(d.Get(names.AttrRole).(string)),
			Rules: expandReplicationRules(ctx, d.Get(names.AttrRule).([]interface{})),
		},
	}

	if v, ok := d.GetOk("token"); ok {
		input.Token = aws.String(v.(string))
	}

	_, err := conn.PutBucketReplication(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Bucket Replication Configuration (%s): %s", d.Id(), err)
	}

	return append(diags, resourceBucketReplicationConfigurationRead(ctx, d, meta)...)
}

func resourceBucketReplicationConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	log.Printf("[DEBUG] Deleting S3 Bucket Replication Configuration: %s", d.Id())
	_, err := conn.DeleteBucketReplication(ctx, &s3.DeleteBucketReplicationInput{
		Bucket: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeReplicationConfigurationNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket Replication Configuration (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return findReplicationConfiguration(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Replication Configuration (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findReplicationConfiguration(ctx context.Context, conn *s3.Client, bucket string) (*types.ReplicationConfiguration, error) {
	input := &s3.GetBucketReplicationInput{
		Bucket: aws.String(bucket),
	}

	output, err := conn.GetBucketReplication(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeReplicationConfigurationNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ReplicationConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ReplicationConfiguration, nil
}

func expandReplicationRules(ctx context.Context, l []interface{}) []types.ReplicationRule {
	var rules []types.ReplicationRule

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		rule := types.ReplicationRule{}

		if v, ok := tfMap["delete_marker_replication"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.DeleteMarkerReplication = expandDeleteMarkerReplication(v)
		}

		if v, ok := tfMap[names.AttrDestination].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.Destination = expandDestination(v)
		}

		if v, ok := tfMap["existing_object_replication"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.ExistingObjectReplication = expandExistingObjectReplication(v)
		}

		if v, ok := tfMap[names.AttrID].(string); ok && v != "" {
			rule.ID = aws.String(v)
		}

		if v, ok := tfMap["source_selection_criteria"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.SourceSelectionCriteria = expandSourceSelectionCriteria(v)
		}

		if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
			rule.Status = types.ReplicationRuleStatus(v)
		}

		// Support the empty filter block in terraform i.e. 'filter {}',
		// which implies the replication rule does not require a specific filter,
		// by expanding the "filter" array even if the first element is nil.
		if v, ok := tfMap[names.AttrFilter].([]interface{}); ok && len(v) > 0 {
			// XML schema V2
			rule.Filter = expandReplicationRuleFilter(ctx, v)
			rule.Priority = aws.Int32(int32(tfMap[names.AttrPriority].(int)))
		} else {
			// XML schema V1
			rule.Prefix = aws.String(tfMap[names.AttrPrefix].(string))
		}

		rules = append(rules, rule)
	}

	return rules
}

func expandDeleteMarkerReplication(l []interface{}) *types.DeleteMarkerReplication {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.DeleteMarkerReplication{}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		result.Status = types.DeleteMarkerReplicationStatus(v)
	}

	return result
}

func expandDestination(l []interface{}) *types.Destination {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.Destination{}

	if v, ok := tfMap["access_control_translation"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.AccessControlTranslation = expandAccessControlTranslation(v)
	}

	if v, ok := tfMap["account"].(string); ok && v != "" {
		result.Account = aws.String(v)
	}

	if v, ok := tfMap[names.AttrBucket].(string); ok && v != "" {
		result.Bucket = aws.String(v)
	}

	if v, ok := tfMap[names.AttrEncryptionConfiguration].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.EncryptionConfiguration = expandEncryptionConfiguration(v)
	}

	if v, ok := tfMap["metrics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.Metrics = expandMetrics(v)
	}

	if v, ok := tfMap["replication_time"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.ReplicationTime = expandReplicationTime(v)
	}

	if v, ok := tfMap[names.AttrStorageClass].(string); ok && v != "" {
		result.StorageClass = types.StorageClass(v)
	}

	return result
}

func expandAccessControlTranslation(l []interface{}) *types.AccessControlTranslation {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.AccessControlTranslation{}

	if v, ok := tfMap[names.AttrOwner].(string); ok && v != "" {
		result.Owner = types.OwnerOverride(v)
	}

	return result
}

func expandEncryptionConfiguration(l []interface{}) *types.EncryptionConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.EncryptionConfiguration{}

	if v, ok := tfMap["replica_kms_key_id"].(string); ok && v != "" {
		result.ReplicaKmsKeyID = aws.String(v)
	}

	return result
}

func expandMetrics(l []interface{}) *types.Metrics {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.Metrics{}

	if v, ok := tfMap["event_threshold"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.EventThreshold = expandReplicationTimeValue(v)
	}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		result.Status = types.MetricsStatus(v)
	}

	return result
}

func expandReplicationTime(l []interface{}) *types.ReplicationTime {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.ReplicationTime{}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		result.Status = types.ReplicationTimeStatus(v)
	}

	if v, ok := tfMap["time"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.Time = expandReplicationTimeValue(v)
	}

	return result
}

func expandReplicationTimeValue(l []interface{}) *types.ReplicationTimeValue {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.ReplicationTimeValue{}

	if v, ok := tfMap["minutes"].(int); ok {
		result.Minutes = aws.Int32(int32(v))
	}

	return result
}

func expandExistingObjectReplication(l []interface{}) *types.ExistingObjectReplication {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.ExistingObjectReplication{}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		result.Status = types.ExistingObjectReplicationStatus(v)
	}

	return result
}

func expandSourceSelectionCriteria(l []interface{}) *types.SourceSelectionCriteria {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.SourceSelectionCriteria{}

	if v, ok := tfMap["replica_modifications"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.ReplicaModifications = expandReplicaModifications(v)
	}

	if v, ok := tfMap["sse_kms_encrypted_objects"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.SseKmsEncryptedObjects = expandSSEKMSEncryptedObjects(v)
	}

	return result
}

func expandReplicaModifications(l []interface{}) *types.ReplicaModifications {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.ReplicaModifications{}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		result.Status = types.ReplicaModificationsStatus(v)
	}

	return result
}

func expandSSEKMSEncryptedObjects(l []interface{}) *types.SseKmsEncryptedObjects {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.SseKmsEncryptedObjects{}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		result.Status = types.SseKmsEncryptedObjectsStatus(v)
	}

	return result
}

func expandReplicationRuleFilter(ctx context.Context, l []interface{}) types.ReplicationRuleFilter {
	if len(l) == 0 || l[0] == nil {
		return &types.ReplicationRuleFilterMemberPrefix{}
	}

	tfMap := l[0].(map[string]interface{})
	var result types.ReplicationRuleFilter

	if v, ok := tfMap["and"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result = expandReplicationRuleFilterMemberAnd(ctx, v)
	}

	if v, ok := tfMap["tag"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result = expandReplicationRuleFilterMemberTag(v)
	}

	// Per AWS S3 API, "A Filter must have exactly one of Prefix, Tag, or And specified";
	// Specifying more than one of the listed parameters results in a MalformedXML error.
	// If a filter is specified as filter { prefix = "" } in Terraform, we should send the prefix value
	// in the API request even if it is an empty value, else Terraform will report non-empty plans.
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/23487
	if v, ok := tfMap[names.AttrPrefix].(string); ok && result == nil {
		result = &types.ReplicationRuleFilterMemberPrefix{
			Value: v,
		}
	}

	return result
}

func expandReplicationRuleFilterMemberAnd(ctx context.Context, l []interface{}) *types.ReplicationRuleFilterMemberAnd {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.ReplicationRuleFilterMemberAnd{
		Value: types.ReplicationRuleAndOperator{},
	}

	if v, ok := tfMap[names.AttrPrefix].(string); ok && v != "" {
		result.Value.Prefix = aws.String(v)
	}

	if v, ok := tfMap[names.AttrTags].(map[string]interface{}); ok && len(v) > 0 {
		tags := Tags(tftags.New(ctx, v).IgnoreAWS())
		if len(tags) > 0 {
			result.Value.Tags = tags
		}
	}

	return result
}

func expandReplicationRuleFilterMemberTag(l []interface{}) *types.ReplicationRuleFilterMemberTag {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.ReplicationRuleFilterMemberTag{
		Value: types.Tag{},
	}

	if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
		result.Value.Key = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		result.Value.Value = aws.String(v)
	}

	return result
}

func flattenReplicationRules(ctx context.Context, rules []types.ReplicationRule) []interface{} {
	if len(rules) == 0 {
		return []interface{}{}
	}

	var results []interface{}

	for _, rule := range rules {
		m := map[string]interface{}{
			names.AttrStatus: rule.Status,
		}

		if rule.DeleteMarkerReplication != nil {
			m["delete_marker_replication"] = flattenDeleteMarkerReplication(rule.DeleteMarkerReplication)
		}

		if rule.Destination != nil {
			m[names.AttrDestination] = flattenDestination(rule.Destination)
		}

		if rule.ExistingObjectReplication != nil {
			m["existing_object_replication"] = flattenExistingObjectReplication(rule.ExistingObjectReplication)
		}

		if rule.Filter != nil {
			m[names.AttrFilter] = flattenReplicationRuleFilter(ctx, rule.Filter)
		}

		if rule.ID != nil {
			m[names.AttrID] = aws.ToString(rule.ID)
		}

		if rule.Prefix != nil {
			m[names.AttrPrefix] = aws.ToString(rule.Prefix)
		}

		if rule.Priority != nil {
			m[names.AttrPriority] = aws.ToInt32(rule.Priority)
		}

		if rule.SourceSelectionCriteria != nil {
			m["source_selection_criteria"] = flattenSourceSelectionCriteria(rule.SourceSelectionCriteria)
		}

		results = append(results, m)
	}

	return results
}

func flattenDeleteMarkerReplication(dmr *types.DeleteMarkerReplication) []interface{} {
	if dmr == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrStatus: dmr.Status,
	}

	return []interface{}{m}
}

func flattenDestination(dest *types.Destination) []interface{} {
	if dest == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrStorageClass: dest.StorageClass,
	}

	if dest.AccessControlTranslation != nil {
		m["access_control_translation"] = flattenAccessControlTranslation(dest.AccessControlTranslation)
	}

	if dest.Account != nil {
		m["account"] = aws.ToString(dest.Account)
	}

	if dest.Bucket != nil {
		m[names.AttrBucket] = aws.ToString(dest.Bucket)
	}

	if dest.EncryptionConfiguration != nil {
		m[names.AttrEncryptionConfiguration] = flattenEncryptionConfiguration(dest.EncryptionConfiguration)
	}

	if dest.Metrics != nil {
		m["metrics"] = flattenMetrics(dest.Metrics)
	}

	if dest.ReplicationTime != nil {
		m["replication_time"] = flattenReplicationReplicationTime(dest.ReplicationTime)
	}

	return []interface{}{m}
}

func flattenAccessControlTranslation(act *types.AccessControlTranslation) []interface{} {
	if act == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrOwner: act.Owner,
	}

	return []interface{}{m}
}

func flattenEncryptionConfiguration(ec *types.EncryptionConfiguration) []interface{} {
	if ec == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if ec.ReplicaKmsKeyID != nil {
		m["replica_kms_key_id"] = aws.ToString(ec.ReplicaKmsKeyID)
	}

	return []interface{}{m}
}

func flattenMetrics(metrics *types.Metrics) []interface{} {
	if metrics == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrStatus: metrics.Status,
	}

	if metrics.EventThreshold != nil {
		m["event_threshold"] = flattenReplicationTimeValue(metrics.EventThreshold)
	}

	return []interface{}{m}
}

func flattenReplicationTimeValue(rtv *types.ReplicationTimeValue) []interface{} {
	if rtv == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"minutes": rtv.Minutes,
	}

	return []interface{}{m}
}

func flattenReplicationReplicationTime(rt *types.ReplicationTime) []interface{} {
	if rt == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrStatus: rt.Status,
	}

	if rt.Time != nil {
		m["time"] = flattenReplicationTimeValue(rt.Time)
	}

	return []interface{}{m}
}

func flattenExistingObjectReplication(eor *types.ExistingObjectReplication) []interface{} {
	if eor == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrStatus: eor.Status,
	}

	return []interface{}{m}
}

func flattenReplicationRuleFilter(ctx context.Context, filter types.ReplicationRuleFilter) []interface{} {
	if filter == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	switch v := filter.(type) {
	case *types.ReplicationRuleFilterMemberAnd:
		m["and"] = flattenReplicationRuleFilterMemberAnd(ctx, v)
	case *types.ReplicationRuleFilterMemberPrefix:
		m[names.AttrPrefix] = v.Value
	case *types.ReplicationRuleFilterMemberTag:
		m["tag"] = flattenReplicationRuleFilterMemberTag(v)
	default:
		return nil
	}

	return []interface{}{m}
}

func flattenReplicationRuleFilterMemberAnd(ctx context.Context, op *types.ReplicationRuleFilterMemberAnd) []interface{} {
	if op == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if v := op.Value.Prefix; v != nil {
		m[names.AttrPrefix] = aws.ToString(v)
	}

	if v := op.Value.Tags; v != nil {
		m[names.AttrTags] = keyValueTags(ctx, v).IgnoreAWS().Map()
	}

	return []interface{}{m}
}

func flattenReplicationRuleFilterMemberTag(op *types.ReplicationRuleFilterMemberTag) []interface{} {
	if op == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if v := op.Value.Key; v != nil {
		m[names.AttrKey] = aws.ToString(v)
	}

	if v := op.Value.Value; v != nil {
		m[names.AttrValue] = aws.ToString(v)
	}

	return []interface{}{m}
}

func flattenSourceSelectionCriteria(ssc *types.SourceSelectionCriteria) []interface{} {
	if ssc == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if ssc.ReplicaModifications != nil {
		m["replica_modifications"] = flattenReplicaModifications(ssc.ReplicaModifications)
	}

	if ssc.SseKmsEncryptedObjects != nil {
		m["sse_kms_encrypted_objects"] = flattenSSEKMSEncryptedObjects(ssc.SseKmsEncryptedObjects)
	}

	return []interface{}{m}
}

func flattenReplicaModifications(rc *types.ReplicaModifications) []interface{} {
	if rc == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrStatus: rc.Status,
	}

	return []interface{}{m}
}

func flattenSSEKMSEncryptedObjects(objects *types.SseKmsEncryptedObjects) []interface{} {
	if objects == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrStatus: objects.Status,
	}

	return []interface{}{m}
}
