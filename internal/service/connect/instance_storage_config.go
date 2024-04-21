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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_connect_instance_storage_config")
func ResourceInstanceStorageConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceStorageConfigCreate,
		ReadWithoutTimeout:   resourceInstanceStorageConfigRead,
		UpdateWithoutTimeout: resourceInstanceStorageConfigUpdate,
		DeleteWithoutTimeout: resourceInstanceStorageConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"resource_type": {
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
									"stream_arn": {
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
												"key_id": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									"prefix": {
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
									"bucket_name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
									"bucket_prefix": {
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
												"key_id": {
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
						"storage_type": {
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

func resourceInstanceStorageConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceId := d.Get("instance_id").(string)
	resourceType := d.Get("resource_type").(string)

	input := &connect.AssociateInstanceStorageConfigInput{
		InstanceId:    aws.String(instanceId),
		ResourceType:  awstypes.InstanceStorageResourceType(resourceType),
		StorageConfig: expandStorageConfig(d.Get("storage_config").([]interface{})),
	}

	log.Printf("[DEBUG] Creating Connect Instance Storage Config %+v", input)
	output, err := conn.AssociateInstanceStorageConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Instance Storage Config for Connect Instance (%s,%s): %s", instanceId, resourceType, err)
	}

	if output == nil || output.AssociationId == nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Instance Storage Config for Connect Instance (%s,%s): empty output", instanceId, resourceType)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", instanceId, aws.ToString(output.AssociationId), resourceType))

	return append(diags, resourceInstanceStorageConfigRead(ctx, d, meta)...)
}

func resourceInstanceStorageConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceId, associationId, resourceType, err := InstanceStorageConfigParseId(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	resp, err := conn.DescribeInstanceStorageConfig(ctx, &connect.DescribeInstanceStorageConfigInput{
		AssociationId: aws.String(associationId),
		InstanceId:    aws.String(instanceId),
		ResourceType:  awstypes.InstanceStorageResourceType(resourceType),
	})

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] Connect Instance Storage Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Instance Storage Config (%s): %s", d.Id(), err)
	}

	if resp == nil || resp.StorageConfig == nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Instance Storage Config (%s): empty response", d.Id())
	}

	storageConfig := resp.StorageConfig

	d.Set("association_id", storageConfig.AssociationId)
	d.Set("instance_id", instanceId)
	d.Set("resource_type", resourceType)

	if err := d.Set("storage_config", flattenStorageConfig(storageConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting storage_config: %s", err)
	}

	return diags
}

func resourceInstanceStorageConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceId, associationId, resourceType, err := InstanceStorageConfigParseId(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &connect.UpdateInstanceStorageConfigInput{
		AssociationId: aws.String(associationId),
		InstanceId:    aws.String(instanceId),
		ResourceType:  awstypes.InstanceStorageResourceType(resourceType),
	}

	if d.HasChange("storage_config") {
		input.StorageConfig = expandStorageConfig(d.Get("storage_config").([]interface{}))
	}

	_, err = conn.UpdateInstanceStorageConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Instance Storage Config (%s): %s", d.Id(), err)
	}

	return append(diags, resourceInstanceStorageConfigRead(ctx, d, meta)...)
}

func resourceInstanceStorageConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceId, associationId, resourceType, err := InstanceStorageConfigParseId(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = conn.DisassociateInstanceStorageConfig(ctx, &connect.DisassociateInstanceStorageConfigInput{
		AssociationId: aws.String(associationId),
		InstanceId:    aws.String(instanceId),
		ResourceType:  awstypes.InstanceStorageResourceType(resourceType),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting InstanceStorageConfig (%s): %s", d.Id(), err)
	}

	return diags
}

func InstanceStorageConfigParseId(id string) (string, string, string, error) {
	parts := strings.SplitN(id, ":", 3)

	if len(parts) < 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceId:associationId:resourceType", id)
	}

	return parts[0], parts[1], parts[2], nil
}

func expandStorageConfig(tfList []interface{}) *awstypes.InstanceStorageConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &awstypes.InstanceStorageConfig{
		StorageType: awstypes.StorageType(tfMap["storage_type"].(string)),
	}

	if v, ok := tfMap["kinesis_firehose_config"].([]interface{}); ok && len(v) > 0 {
		result.KinesisFirehoseConfig = expandKinesisFirehoseConfig(v)
	}

	if v, ok := tfMap["kinesis_stream_config"].([]interface{}); ok && len(v) > 0 {
		result.KinesisStreamConfig = expandKinesisStreamConfig(v)
	}

	if v, ok := tfMap["kinesis_video_stream_config"].([]interface{}); ok && len(v) > 0 {
		result.KinesisVideoStreamConfig = expandKinesisVideoStreamConfig(v)
	}

	if v, ok := tfMap["s3_config"].([]interface{}); ok && len(v) > 0 {
		result.S3Config = exapandS3Config(v)
	}

	return result
}

func expandKinesisFirehoseConfig(tfList []interface{}) *awstypes.KinesisFirehoseConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &awstypes.KinesisFirehoseConfig{
		FirehoseArn: aws.String(tfMap["firehose_arn"].(string)),
	}

	return result
}

func expandKinesisStreamConfig(tfList []interface{}) *awstypes.KinesisStreamConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &awstypes.KinesisStreamConfig{
		StreamArn: aws.String(tfMap["stream_arn"].(string)),
	}

	return result
}

func expandKinesisVideoStreamConfig(tfList []interface{}) *awstypes.KinesisVideoStreamConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &awstypes.KinesisVideoStreamConfig{
		EncryptionConfig:     expandEncryptionConfig(tfMap["encryption_config"].([]interface{})),
		Prefix:               aws.String(tfMap["prefix"].(string)),
		RetentionPeriodHours: int32(tfMap["retention_period_hours"].(int)),
	}

	return result
}

func exapandS3Config(tfList []interface{}) *awstypes.S3Config {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &awstypes.S3Config{
		BucketName:   aws.String(tfMap["bucket_name"].(string)),
		BucketPrefix: aws.String(tfMap["bucket_prefix"].(string)),
	}

	if v, ok := tfMap["encryption_config"].([]interface{}); ok && len(v) > 0 {
		result.EncryptionConfig = expandEncryptionConfig(v)
	}

	return result
}

func expandEncryptionConfig(tfList []interface{}) *awstypes.EncryptionConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &awstypes.EncryptionConfig{
		EncryptionType: awstypes.EncryptionType(tfMap["encryption_type"].(string)),
		KeyId:          aws.String(tfMap["key_id"].(string)),
	}

	return result
}

func flattenStorageConfig(apiObject *awstypes.InstanceStorageConfig) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"storage_type": string(apiObject.StorageType),
	}

	if v := apiObject.KinesisFirehoseConfig; v != nil {
		values["kinesis_firehose_config"] = flattenKinesisFirehoseConfig(v)
	}

	if v := apiObject.KinesisStreamConfig; v != nil {
		values["kinesis_stream_config"] = flattenKinesisStreamConfig(v)
	}

	if v := apiObject.KinesisVideoStreamConfig; v != nil {
		values["kinesis_video_stream_config"] = flattenKinesisVideoStreamConfig(v)
	}

	if v := apiObject.S3Config; v != nil {
		values["s3_config"] = flattenS3Config(v)
	}

	return []interface{}{values}
}

func flattenKinesisFirehoseConfig(apiObject *awstypes.KinesisFirehoseConfig) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"firehose_arn": aws.ToString(apiObject.FirehoseArn),
	}

	return []interface{}{values}
}

func flattenKinesisStreamConfig(apiObject *awstypes.KinesisStreamConfig) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"stream_arn": aws.ToString(apiObject.StreamArn),
	}

	return []interface{}{values}
}

func flattenKinesisVideoStreamConfig(apiObject *awstypes.KinesisVideoStreamConfig) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"encryption_config": flattenEncryptionConfig(apiObject.EncryptionConfig),
		// API returns <prefix>-connect-<connect_instance_alias>-contact-
		// DiffSuppressFunc used
		"prefix":                 aws.ToString(apiObject.Prefix),
		"retention_period_hours": apiObject.RetentionPeriodHours,
	}

	return []interface{}{values}
}

func flattenS3Config(apiObject *awstypes.S3Config) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"bucket_name":   aws.ToString(apiObject.BucketName),
		"bucket_prefix": aws.ToString(apiObject.BucketPrefix),
	}

	if v := apiObject.EncryptionConfig; v != nil {
		values["encryption_config"] = flattenEncryptionConfig(v)
	}

	return []interface{}{values}
}

func flattenEncryptionConfig(apiObject *awstypes.EncryptionConfig) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"encryption_type": string(apiObject.EncryptionType),
		"key_id":          aws.ToString(apiObject.KeyId),
	}

	return []interface{}{values}
}
