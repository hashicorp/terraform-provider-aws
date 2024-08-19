// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(connect.InstanceStorageResourceType_Values(), false),
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
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(connect.EncryptionType_Values(), false),
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
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(connect.EncryptionType_Values(), false),
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
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(connect.StorageType_Values(), false),
						},
					},
				},
			},
		},
	}
}

func resourceInstanceStorageConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceId := d.Get(names.AttrInstanceID).(string)
	resourceType := d.Get(names.AttrResourceType).(string)

	input := &connect.AssociateInstanceStorageConfigInput{
		InstanceId:    aws.String(instanceId),
		ResourceType:  aws.String(resourceType),
		StorageConfig: expandStorageConfig(d.Get("storage_config").([]interface{})),
	}

	log.Printf("[DEBUG] Creating Connect Instance Storage Config %s", input)
	output, err := conn.AssociateInstanceStorageConfigWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Instance Storage Config for Connect Instance (%s,%s): %s", instanceId, resourceType, err)
	}

	if output == nil || output.AssociationId == nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Instance Storage Config for Connect Instance (%s,%s): empty output", instanceId, resourceType)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", instanceId, aws.StringValue(output.AssociationId), resourceType))

	return append(diags, resourceInstanceStorageConfigRead(ctx, d, meta)...)
}

func resourceInstanceStorageConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceId, associationId, resourceType, err := InstanceStorageConfigParseId(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	resp, err := conn.DescribeInstanceStorageConfigWithContext(ctx, &connect.DescribeInstanceStorageConfigInput{
		AssociationId: aws.String(associationId),
		InstanceId:    aws.String(instanceId),
		ResourceType:  aws.String(resourceType),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
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

	d.Set(names.AttrAssociationID, storageConfig.AssociationId)
	d.Set(names.AttrInstanceID, instanceId)
	d.Set(names.AttrResourceType, resourceType)

	if err := d.Set("storage_config", flattenStorageConfig(storageConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting storage_config: %s", err)
	}

	return diags
}

func resourceInstanceStorageConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceId, associationId, resourceType, err := InstanceStorageConfigParseId(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &connect.UpdateInstanceStorageConfigInput{
		AssociationId: aws.String(associationId),
		InstanceId:    aws.String(instanceId),
		ResourceType:  aws.String(resourceType),
	}

	if d.HasChange("storage_config") {
		input.StorageConfig = expandStorageConfig(d.Get("storage_config").([]interface{}))
	}

	_, err = conn.UpdateInstanceStorageConfigWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Instance Storage Config (%s): %s", d.Id(), err)
	}

	return append(diags, resourceInstanceStorageConfigRead(ctx, d, meta)...)
}

func resourceInstanceStorageConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceId, associationId, resourceType, err := InstanceStorageConfigParseId(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = conn.DisassociateInstanceStorageConfigWithContext(ctx, &connect.DisassociateInstanceStorageConfigInput{
		AssociationId: aws.String(associationId),
		InstanceId:    aws.String(instanceId),
		ResourceType:  aws.String(resourceType),
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

func expandStorageConfig(tfList []interface{}) *connect.InstanceStorageConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &connect.InstanceStorageConfig{
		StorageType: aws.String(tfMap[names.AttrStorageType].(string)),
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

func expandKinesisFirehoseConfig(tfList []interface{}) *connect.KinesisFirehoseConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &connect.KinesisFirehoseConfig{
		FirehoseArn: aws.String(tfMap["firehose_arn"].(string)),
	}

	return result
}

func expandKinesisStreamConfig(tfList []interface{}) *connect.KinesisStreamConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &connect.KinesisStreamConfig{
		StreamArn: aws.String(tfMap[names.AttrStreamARN].(string)),
	}

	return result
}

func expandKinesisVideoStreamConfig(tfList []interface{}) *connect.KinesisVideoStreamConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &connect.KinesisVideoStreamConfig{
		EncryptionConfig:     expandEncryptionConfig(tfMap["encryption_config"].([]interface{})),
		Prefix:               aws.String(tfMap[names.AttrPrefix].(string)),
		RetentionPeriodHours: aws.Int64(int64(tfMap["retention_period_hours"].(int))),
	}

	return result
}

func exapandS3Config(tfList []interface{}) *connect.S3Config {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &connect.S3Config{
		BucketName:   aws.String(tfMap[names.AttrBucketName].(string)),
		BucketPrefix: aws.String(tfMap[names.AttrBucketPrefix].(string)),
	}

	if v, ok := tfMap["encryption_config"].([]interface{}); ok && len(v) > 0 {
		result.EncryptionConfig = expandEncryptionConfig(v)
	}

	return result
}

func expandEncryptionConfig(tfList []interface{}) *connect.EncryptionConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &connect.EncryptionConfig{
		EncryptionType: aws.String(tfMap["encryption_type"].(string)),
		KeyId:          aws.String(tfMap[names.AttrKeyID].(string)),
	}

	return result
}

func flattenStorageConfig(apiObject *connect.InstanceStorageConfig) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		names.AttrStorageType: aws.StringValue(apiObject.StorageType),
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

func flattenKinesisFirehoseConfig(apiObject *connect.KinesisFirehoseConfig) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"firehose_arn": aws.StringValue(apiObject.FirehoseArn),
	}

	return []interface{}{values}
}

func flattenKinesisStreamConfig(apiObject *connect.KinesisStreamConfig) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		names.AttrStreamARN: aws.StringValue(apiObject.StreamArn),
	}

	return []interface{}{values}
}

func flattenKinesisVideoStreamConfig(apiObject *connect.KinesisVideoStreamConfig) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"encryption_config": flattenEncryptionConfig(apiObject.EncryptionConfig),
		// API returns <prefix>-connect-<connect_instance_alias>-contact-
		// DiffSuppressFunc used
		names.AttrPrefix:         aws.StringValue(apiObject.Prefix),
		"retention_period_hours": aws.Int64Value(apiObject.RetentionPeriodHours),
	}

	return []interface{}{values}
}

func flattenS3Config(apiObject *connect.S3Config) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		names.AttrBucketName:   aws.StringValue(apiObject.BucketName),
		names.AttrBucketPrefix: aws.StringValue(apiObject.BucketPrefix),
	}

	if v := apiObject.EncryptionConfig; v != nil {
		values["encryption_config"] = flattenEncryptionConfig(v)
	}

	return []interface{}{values}
}

func flattenEncryptionConfig(apiObject *connect.EncryptionConfig) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"encryption_type": aws.StringValue(apiObject.EncryptionType),
		names.AttrKeyID:   aws.StringValue(apiObject.KeyId),
	}

	return []interface{}{values}
}
