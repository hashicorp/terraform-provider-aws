// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_connect_instance_storage_config", name="Instance Storage Config")
func resourceInstanceStorageConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceStorageConfigCreate,
		ReadWithoutTimeout:   resourceInstanceStorageConfigRead,
		UpdateWithoutTimeout: resourceInstanceStorageConfigUpdate,
		DeleteWithoutTimeout: resourceInstanceStorageConfigDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAssociationID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrInstanceID: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			names.AttrResourceType: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.InstanceStorageResourceType](),
			},
			"storage_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kinesis_firehose_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"firehose_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"kinesis_stream_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrStreamARN: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"kinesis_video_stream_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"encryption_config": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"encryption_type": {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.EncryptionType](),
												},
												names.AttrKeyID: {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									names.AttrPrefix: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 128),
										// API returns <prefix>-connect-<connect_instance_alias>-contact-
										DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
											// API returns <prefix>-connect-<connect_instance_alias>-contact-
											// case 1: API appends to prefix. User-defined string (old) is prefix of API-returned string (new). Check non-empty old in resoure creation scenario
											// case 2: after setting API-returned string.  User-defined string (new) is prefix of API-returned string (old)
											// case 3: update for other arguments that still require the prefix to be sent in the request
											return (strings.HasPrefix(new, old) && old != "") || (strings.HasPrefix(old, new) && !d.HasChange("storage_config.0.kinesis_video_stream_config.0.encryption_config") && !d.HasChange("storage_config.0.kinesis_video_stream_config.0.retention_period_hours"))
										},
									},
									"retention_period_hours": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(0, 87600),
									},
								},
							},
						},
						"s3_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrBucketName: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
									names.AttrBucketPrefix: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
									"encryption_config": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"encryption_type": {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.EncryptionType](),
												},
												names.AttrKeyID: {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
								},
							},
						},
						names.AttrStorageType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.StorageType](),
						},
					},
				},
			},
		},
	}
}

func resourceInstanceStorageConfigCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	resourceType := awstypes.InstanceStorageResourceType(d.Get(names.AttrResourceType).(string))
	input := &connect.AssociateInstanceStorageConfigInput{
		InstanceId:    aws.String(instanceID),
		ResourceType:  resourceType,
		StorageConfig: expandInstanceStorageConfig(d.Get("storage_config").([]any)),
	}

	output, err := conn.AssociateInstanceStorageConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Instance (%s) Storage Config (%s): %s", instanceID, resourceType, err)
	}

	id := instanceStorageConfigCreateResourceID(instanceID, aws.ToString(output.AssociationId), resourceType)
	d.SetId(id)

	return append(diags, resourceInstanceStorageConfigRead(ctx, d, meta)...)
}

func resourceInstanceStorageConfigRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, associationID, resourceType, err := instanceStorageConfigParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	storageConfig, err := findInstanceStorageConfigByThreePartKey(ctx, conn, instanceID, associationID, resourceType)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect Instance Storage Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Instance Storage Config (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAssociationID, storageConfig.AssociationId)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set(names.AttrResourceType, resourceType)
	if err := d.Set("storage_config", flattenStorageConfig(storageConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting storage_config: %s", err)
	}

	return diags
}

func resourceInstanceStorageConfigUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, associationID, resourceType, err := instanceStorageConfigParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &connect.UpdateInstanceStorageConfigInput{
		AssociationId: aws.String(associationID),
		InstanceId:    aws.String(instanceID),
		ResourceType:  resourceType,
	}

	if d.HasChange("storage_config") {
		input.StorageConfig = expandInstanceStorageConfig(d.Get("storage_config").([]any))
	}

	_, err = conn.UpdateInstanceStorageConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Connect Instance Storage Config (%s): %s", d.Id(), err)
	}

	return append(diags, resourceInstanceStorageConfigRead(ctx, d, meta)...)
}

func resourceInstanceStorageConfigDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, associationID, resourceType, err := instanceStorageConfigParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Connect Instance Storage Config: %s", d.Id())
	input := connect.DisassociateInstanceStorageConfigInput{
		AssociationId: aws.String(associationID),
		InstanceId:    aws.String(instanceID),
		ResourceType:  resourceType,
	}
	_, err = conn.DisassociateInstanceStorageConfig(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Connect Instance Storage Config (%s): %s", d.Id(), err)
	}

	return diags
}

const instanceStorageConfigResourceIDSeparator = ":"

func instanceStorageConfigCreateResourceID(instanceID, associationID string, resourceType awstypes.InstanceStorageResourceType) string {
	parts := []string{instanceID, associationID, string(resourceType)} // nosemgrep:ci.typed-enum-conversion
	id := strings.Join(parts, instanceStorageConfigResourceIDSeparator)

	return id
}

func instanceStorageConfigParseResourceID(id string) (string, string, awstypes.InstanceStorageResourceType, error) {
	parts := strings.SplitN(id, instanceStorageConfigResourceIDSeparator, 3)

	if len(parts) < 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected instanceID%[2]sassociationID%[2]sresourceType", id, instanceStorageConfigResourceIDSeparator)
	}

	return parts[0], parts[1], awstypes.InstanceStorageResourceType(parts[2]), nil
}

func findInstanceStorageConfigByThreePartKey(ctx context.Context, conn *connect.Client, instanceID, associationID string, resourceType awstypes.InstanceStorageResourceType) (*awstypes.InstanceStorageConfig, error) {
	input := &connect.DescribeInstanceStorageConfigInput{
		AssociationId: aws.String(associationID),
		InstanceId:    aws.String(instanceID),
		ResourceType:  resourceType,
	}

	return findInstanceStorageConfig(ctx, conn, input)
}

func findInstanceStorageConfig(ctx context.Context, conn *connect.Client, input *connect.DescribeInstanceStorageConfigInput) (*awstypes.InstanceStorageConfig, error) {
	output, err := conn.DescribeInstanceStorageConfig(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.StorageConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.StorageConfig, nil
}

func expandInstanceStorageConfig(tfList []any) *awstypes.InstanceStorageConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.InstanceStorageConfig{
		StorageType: awstypes.StorageType(tfMap[names.AttrStorageType].(string)),
	}

	if v, ok := tfMap["kinesis_firehose_config"].([]any); ok && len(v) > 0 {
		apiObject.KinesisFirehoseConfig = expandKinesisFirehoseConfig(v)
	}

	if v, ok := tfMap["kinesis_stream_config"].([]any); ok && len(v) > 0 {
		apiObject.KinesisStreamConfig = expandKinesisStreamConfig(v)
	}

	if v, ok := tfMap["kinesis_video_stream_config"].([]any); ok && len(v) > 0 {
		apiObject.KinesisVideoStreamConfig = expandKinesisVideoStreamConfig(v)
	}

	if v, ok := tfMap["s3_config"].([]any); ok && len(v) > 0 {
		apiObject.S3Config = exapandS3Config(v)
	}

	return apiObject
}

func expandKinesisFirehoseConfig(tfList []any) *awstypes.KinesisFirehoseConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.KinesisFirehoseConfig{
		FirehoseArn: aws.String(tfMap["firehose_arn"].(string)),
	}

	return apiObject
}

func expandKinesisStreamConfig(tfList []any) *awstypes.KinesisStreamConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.KinesisStreamConfig{
		StreamArn: aws.String(tfMap[names.AttrStreamARN].(string)),
	}

	return apiObject
}

func expandKinesisVideoStreamConfig(tfList []any) *awstypes.KinesisVideoStreamConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.KinesisVideoStreamConfig{
		EncryptionConfig:     expandEncryptionConfig(tfMap["encryption_config"].([]any)),
		Prefix:               aws.String(tfMap[names.AttrPrefix].(string)),
		RetentionPeriodHours: int32(tfMap["retention_period_hours"].(int)),
	}

	return apiObject
}

func exapandS3Config(tfList []any) *awstypes.S3Config {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.S3Config{
		BucketName:   aws.String(tfMap[names.AttrBucketName].(string)),
		BucketPrefix: aws.String(tfMap[names.AttrBucketPrefix].(string)),
	}

	if v, ok := tfMap["encryption_config"].([]any); ok && len(v) > 0 {
		apiObject.EncryptionConfig = expandEncryptionConfig(v)
	}

	return apiObject
}

func expandEncryptionConfig(tfList []any) *awstypes.EncryptionConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.EncryptionConfig{
		EncryptionType: awstypes.EncryptionType(tfMap["encryption_type"].(string)),
		KeyId:          aws.String(tfMap[names.AttrKeyID].(string)),
	}

	return apiObject
}

func flattenStorageConfig(apiObject *awstypes.InstanceStorageConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrStorageType: apiObject.StorageType,
	}

	if v := apiObject.KinesisFirehoseConfig; v != nil {
		tfMap["kinesis_firehose_config"] = flattenKinesisFirehoseConfig(v)
	}

	if v := apiObject.KinesisStreamConfig; v != nil {
		tfMap["kinesis_stream_config"] = flattenKinesisStreamConfig(v)
	}

	if v := apiObject.KinesisVideoStreamConfig; v != nil {
		tfMap["kinesis_video_stream_config"] = flattenKinesisVideoStreamConfig(v)
	}

	if v := apiObject.S3Config; v != nil {
		tfMap["s3_config"] = flattenS3Config(v)
	}

	return []any{tfMap}
}

func flattenKinesisFirehoseConfig(apiObject *awstypes.KinesisFirehoseConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"firehose_arn": aws.ToString(apiObject.FirehoseArn),
	}

	return []any{tfMap}
}

func flattenKinesisStreamConfig(apiObject *awstypes.KinesisStreamConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrStreamARN: aws.ToString(apiObject.StreamArn),
	}

	return []any{tfMap}
}

func flattenKinesisVideoStreamConfig(apiObject *awstypes.KinesisVideoStreamConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"encryption_config": flattenEncryptionConfig(apiObject.EncryptionConfig),
		// API returns <prefix>-connect-<connect_instance_alias>-contact-
		// DiffSuppressFunc used
		names.AttrPrefix:         aws.ToString(apiObject.Prefix),
		"retention_period_hours": apiObject.RetentionPeriodHours,
	}

	return []any{tfMap}
}

func flattenS3Config(apiObject *awstypes.S3Config) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrBucketName:   aws.ToString(apiObject.BucketName),
		names.AttrBucketPrefix: aws.ToString(apiObject.BucketPrefix),
	}

	if v := apiObject.EncryptionConfig; v != nil {
		tfMap["encryption_config"] = flattenEncryptionConfig(v)
	}

	return []any{tfMap}
}

func flattenEncryptionConfig(apiObject *awstypes.EncryptionConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"encryption_type": apiObject.EncryptionType,
		names.AttrKeyID:   aws.ToString(apiObject.KeyId),
	}

	return []any{tfMap}
}
