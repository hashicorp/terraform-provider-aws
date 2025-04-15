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
							Deprecated:   "prefix is deprecated. Use filter instead.",
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

func resourceBucketReplicationConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Get(names.AttrBucket).(string)
	if isDirectoryBucket(bucket) {
		conn = meta.(*conns.AWSClient).S3ExpressClient(ctx)
	}
	input := &s3.PutBucketReplicationInput{
		Bucket: aws.String(bucket),
		ReplicationConfiguration: &types.ReplicationConfiguration{
			Role:  aws.String(d.Get(names.AttrRole).(string)),
			Rules: expandReplicationRules(ctx, d.Get(names.AttrRule).([]any)),
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

	_, err = tfresource.RetryWhenNotFound(ctx, bucketPropagationTimeout, func() (any, error) {
		return findReplicationConfiguration(ctx, conn, bucket)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Replication Configuration (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceBucketReplicationConfigurationRead(ctx, d, meta)...)
}

func resourceBucketReplicationConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Id()
	if isDirectoryBucket(bucket) {
		conn = meta.(*conns.AWSClient).S3ExpressClient(ctx)
	}

	rc, err := findReplicationConfiguration(ctx, conn, bucket)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket Replication Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket Replication Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrBucket, bucket)
	d.Set(names.AttrRole, rc.Role)
	if err := d.Set(names.AttrRule, flattenReplicationRules(ctx, rc.Rules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rule: %s", err)
	}

	return diags
}

func resourceBucketReplicationConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Id()
	if isDirectoryBucket(bucket) {
		conn = meta.(*conns.AWSClient).S3ExpressClient(ctx)
	}

	input := &s3.PutBucketReplicationInput{
		Bucket: aws.String(bucket),
		ReplicationConfiguration: &types.ReplicationConfiguration{
			Role:  aws.String(d.Get(names.AttrRole).(string)),
			Rules: expandReplicationRules(ctx, d.Get(names.AttrRule).([]any)),
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

func resourceBucketReplicationConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Id()
	if isDirectoryBucket(bucket) {
		conn = meta.(*conns.AWSClient).S3ExpressClient(ctx)
	}

	log.Printf("[DEBUG] Deleting S3 Bucket Replication Configuration: %s", d.Id())
	_, err := conn.DeleteBucketReplication(ctx, &s3.DeleteBucketReplicationInput{
		Bucket: aws.String(bucket),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeReplicationConfigurationNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket Replication Configuration (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, bucketPropagationTimeout, func() (any, error) {
		return findReplicationConfiguration(ctx, conn, bucket)
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

func expandReplicationRules(ctx context.Context, tfList []any) []types.ReplicationRule {
	var apiObjects []types.ReplicationRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := types.ReplicationRule{}

		if v, ok := tfMap["delete_marker_replication"].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.DeleteMarkerReplication = expandDeleteMarkerReplication(v)
		}

		if v, ok := tfMap[names.AttrDestination].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.Destination = expandDestination(v)
		}

		if v, ok := tfMap["existing_object_replication"].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.ExistingObjectReplication = expandExistingObjectReplication(v)
		}

		if v, ok := tfMap[names.AttrID].(string); ok && v != "" {
			apiObject.ID = aws.String(v)
		}

		if v, ok := tfMap["source_selection_criteria"].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.SourceSelectionCriteria = expandSourceSelectionCriteria(v)
		}

		if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
			apiObject.Status = types.ReplicationRuleStatus(v)
		}

		// Support the empty filter block in terraform i.e. 'filter {}',
		// which implies the replication rule does not require a specific filter,
		// by expanding the "filter" array even if the first element is nil.
		if v, ok := tfMap[names.AttrFilter].([]any); ok && len(v) > 0 {
			// XML schema V2
			apiObject.Filter = expandReplicationRuleFilter(ctx, v)
			apiObject.Priority = aws.Int32(int32(tfMap[names.AttrPriority].(int)))
		} else {
			// XML schema V1
			apiObject.Prefix = aws.String(tfMap[names.AttrPrefix].(string))
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandDeleteMarkerReplication(tfList []any) *types.DeleteMarkerReplication {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.DeleteMarkerReplication{}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		apiObject.Status = types.DeleteMarkerReplicationStatus(v)
	}

	return apiObject
}

func expandDestination(tfList []any) *types.Destination {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.Destination{}

	if v, ok := tfMap["access_control_translation"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.AccessControlTranslation = expandAccessControlTranslation(v)
	}

	if v, ok := tfMap["account"].(string); ok && v != "" {
		apiObject.Account = aws.String(v)
	}

	if v, ok := tfMap[names.AttrBucket].(string); ok && v != "" {
		apiObject.Bucket = aws.String(v)
	}

	if v, ok := tfMap[names.AttrEncryptionConfiguration].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.EncryptionConfiguration = expandEncryptionConfiguration(v)
	}

	if v, ok := tfMap["metrics"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.Metrics = expandMetrics(v)
	}

	if v, ok := tfMap["replication_time"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.ReplicationTime = expandReplicationTime(v)
	}

	if v, ok := tfMap[names.AttrStorageClass].(string); ok && v != "" {
		apiObject.StorageClass = types.StorageClass(v)
	}

	return apiObject
}

func expandAccessControlTranslation(tfList []any) *types.AccessControlTranslation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.AccessControlTranslation{}

	if v, ok := tfMap[names.AttrOwner].(string); ok && v != "" {
		apiObject.Owner = types.OwnerOverride(v)
	}

	return apiObject
}

func expandEncryptionConfiguration(tfList []any) *types.EncryptionConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.EncryptionConfiguration{}

	if v, ok := tfMap["replica_kms_key_id"].(string); ok && v != "" {
		apiObject.ReplicaKmsKeyID = aws.String(v)
	}

	return apiObject
}

func expandMetrics(tfList []any) *types.Metrics {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.Metrics{}

	if v, ok := tfMap["event_threshold"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.EventThreshold = expandReplicationTimeValue(v)
	}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		apiObject.Status = types.MetricsStatus(v)
	}

	return apiObject
}

func expandReplicationTime(tfList []any) *types.ReplicationTime {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.ReplicationTime{}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		apiObject.Status = types.ReplicationTimeStatus(v)
	}

	if v, ok := tfMap["time"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.Time = expandReplicationTimeValue(v)
	}

	return apiObject
}

func expandReplicationTimeValue(tfList []any) *types.ReplicationTimeValue {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.ReplicationTimeValue{}

	if v, ok := tfMap["minutes"].(int); ok {
		apiObject.Minutes = aws.Int32(int32(v))
	}

	return apiObject
}

func expandExistingObjectReplication(tfList []any) *types.ExistingObjectReplication {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.ExistingObjectReplication{}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		apiObject.Status = types.ExistingObjectReplicationStatus(v)
	}

	return apiObject
}

func expandSourceSelectionCriteria(tfList []any) *types.SourceSelectionCriteria {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.SourceSelectionCriteria{}

	if v, ok := tfMap["replica_modifications"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.ReplicaModifications = expandReplicaModifications(v)
	}

	if v, ok := tfMap["sse_kms_encrypted_objects"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.SseKmsEncryptedObjects = expandSSEKMSEncryptedObjects(v)
	}

	return apiObject
}

func expandReplicaModifications(tfList []any) *types.ReplicaModifications {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.ReplicaModifications{}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		apiObject.Status = types.ReplicaModificationsStatus(v)
	}

	return apiObject
}

func expandSSEKMSEncryptedObjects(tfList []any) *types.SseKmsEncryptedObjects {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.SseKmsEncryptedObjects{}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		apiObject.Status = types.SseKmsEncryptedObjectsStatus(v)
	}

	return apiObject
}

func expandReplicationRuleFilter(ctx context.Context, tfList []any) *types.ReplicationRuleFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return &types.ReplicationRuleFilter{}
	}

	tfMap := tfList[0].(map[string]any)
	var apiObject *types.ReplicationRuleFilter

	if v, ok := tfMap["and"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject = &types.ReplicationRuleFilter{
			And: expandReplicationRuleAndOperator(ctx, v),
		}
	}

	if v, ok := tfMap["tag"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject = &types.ReplicationRuleFilter{
			Tag: expandTag(v[0].(map[string]any)),
		}
	}

	// Per AWS S3 API, "A Filter must have exactly one of Prefix, Tag, or And specified";
	// Specifying more than one of the listed parameters results in a MalformedXML error.
	// If a filter is specified as filter { prefix = "" } in Terraform, we should send the prefix value
	// in the API request even if it is an empty value, else Terraform will report non-empty plans.
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/23487
	if v, ok := tfMap[names.AttrPrefix].(string); ok && apiObject == nil {
		apiObject = &types.ReplicationRuleFilter{
			Prefix: aws.String(v),
		}
	}

	return apiObject
}

func expandReplicationRuleAndOperator(ctx context.Context, tfList []any) *types.ReplicationRuleAndOperator {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.ReplicationRuleAndOperator{}

	if v, ok := tfMap[names.AttrPrefix].(string); ok && v != "" {
		apiObject.Prefix = aws.String(v)
	}

	if v, ok := tfMap[names.AttrTags].(map[string]any); ok && len(v) > 0 {
		if tags := Tags(tftags.New(ctx, v).IgnoreAWS()); len(tags) > 0 {
			apiObject.Tags = tags
		}
	}

	return apiObject
}

func expandTag(tfMap map[string]any) *types.Tag {
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

func flattenReplicationRules(ctx context.Context, apiObjects []types.ReplicationRule) []any {
	if len(apiObjects) == 0 {
		return []any{}
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			names.AttrStatus: apiObject.Status,
		}

		if apiObject.DeleteMarkerReplication != nil {
			tfMap["delete_marker_replication"] = flattenDeleteMarkerReplication(apiObject.DeleteMarkerReplication)
		}

		if apiObject.Destination != nil {
			tfMap[names.AttrDestination] = flattenDestination(apiObject.Destination)
		}

		if apiObject.ExistingObjectReplication != nil {
			tfMap["existing_object_replication"] = flattenExistingObjectReplication(apiObject.ExistingObjectReplication)
		}

		if apiObject.Filter != nil {
			tfMap[names.AttrFilter] = flattenReplicationRuleFilter(ctx, apiObject.Filter)
		}

		if apiObject.ID != nil {
			tfMap[names.AttrID] = aws.ToString(apiObject.ID)
		}

		if apiObject.Prefix != nil {
			tfMap[names.AttrPrefix] = aws.ToString(apiObject.Prefix)
		}

		if apiObject.Priority != nil {
			tfMap[names.AttrPriority] = aws.ToInt32(apiObject.Priority)
		}

		if apiObject.SourceSelectionCriteria != nil {
			tfMap["source_selection_criteria"] = flattenSourceSelectionCriteria(apiObject.SourceSelectionCriteria)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDeleteMarkerReplication(apiObject *types.DeleteMarkerReplication) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrStatus: apiObject.Status,
	}

	return []any{tfMap}
}

func flattenDestination(apiObject *types.Destination) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrStorageClass: apiObject.StorageClass,
	}

	if apiObject.AccessControlTranslation != nil {
		tfMap["access_control_translation"] = flattenAccessControlTranslation(apiObject.AccessControlTranslation)
	}

	if apiObject.Account != nil {
		tfMap["account"] = aws.ToString(apiObject.Account)
	}

	if apiObject.Bucket != nil {
		tfMap[names.AttrBucket] = aws.ToString(apiObject.Bucket)
	}

	if apiObject.EncryptionConfiguration != nil {
		tfMap[names.AttrEncryptionConfiguration] = flattenEncryptionConfiguration(apiObject.EncryptionConfiguration)
	}

	if apiObject.Metrics != nil {
		tfMap["metrics"] = flattenMetrics(apiObject.Metrics)
	}

	if apiObject.ReplicationTime != nil {
		tfMap["replication_time"] = flattenReplicationReplicationTime(apiObject.ReplicationTime)
	}

	return []any{tfMap}
}

func flattenAccessControlTranslation(apiObject *types.AccessControlTranslation) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrOwner: apiObject.Owner,
	}

	return []any{tfMap}
}

func flattenEncryptionConfiguration(apiObject *types.EncryptionConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := make(map[string]any)

	if apiObject.ReplicaKmsKeyID != nil {
		tfMap["replica_kms_key_id"] = aws.ToString(apiObject.ReplicaKmsKeyID)
	}

	return []any{tfMap}
}

func flattenMetrics(apiObject *types.Metrics) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrStatus: apiObject.Status,
	}

	if apiObject.EventThreshold != nil {
		tfMap["event_threshold"] = flattenReplicationTimeValue(apiObject.EventThreshold)
	}

	return []any{tfMap}
}

func flattenReplicationTimeValue(apiObject *types.ReplicationTimeValue) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"minutes": aws.ToInt32(apiObject.Minutes),
	}

	return []any{tfMap}
}

func flattenReplicationReplicationTime(apiObject *types.ReplicationTime) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrStatus: apiObject.Status,
	}

	if apiObject.Time != nil {
		tfMap["time"] = flattenReplicationTimeValue(apiObject.Time)
	}

	return []any{tfMap}
}

func flattenExistingObjectReplication(apiObject *types.ExistingObjectReplication) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrStatus: apiObject.Status,
	}

	return []any{tfMap}
}

func flattenReplicationRuleFilter(ctx context.Context, apiObject *types.ReplicationRuleFilter) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := make(map[string]any)

	if v := apiObject.And; v != nil {
		tfMap["and"] = flattenReplicationRuleAndOperator(ctx, v)
	}

	if v := apiObject.Prefix; v != nil {
		tfMap[names.AttrPrefix] = aws.ToString(v)
	}

	if v := apiObject.Tag; v != nil {
		tfMap["tag"] = flattenTag(v)
	}

	return []any{tfMap}
}

func flattenReplicationRuleAndOperator(ctx context.Context, apiObject *types.ReplicationRuleAndOperator) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := make(map[string]any)

	if v := apiObject.Prefix; v != nil {
		tfMap[names.AttrPrefix] = aws.ToString(v)
	}

	if v := apiObject.Tags; v != nil {
		tfMap[names.AttrTags] = keyValueTags(ctx, v).IgnoreAWS().Map()
	}

	return []any{tfMap}
}

func flattenSourceSelectionCriteria(apiObject *types.SourceSelectionCriteria) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := make(map[string]any)

	if apiObject.ReplicaModifications != nil {
		tfMap["replica_modifications"] = flattenReplicaModifications(apiObject.ReplicaModifications)
	}

	if apiObject.SseKmsEncryptedObjects != nil {
		tfMap["sse_kms_encrypted_objects"] = flattenSSEKMSEncryptedObjects(apiObject.SseKmsEncryptedObjects)
	}

	return []any{tfMap}
}

func flattenReplicaModifications(apiObject *types.ReplicaModifications) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrStatus: apiObject.Status,
	}

	return []any{tfMap}
}

func flattenSSEKMSEncryptedObjects(apiObject *types.SseKmsEncryptedObjects) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrStatus: apiObject.Status,
	}

	return []any{tfMap}
}

func flattenTag(apiObject *types.Tag) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]any)

	if v := apiObject.Key; v != nil {
		tfMap[names.AttrKey] = aws.ToString(v)
	}

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = aws.ToString(v)
	}

	return []any{tfMap}
}
