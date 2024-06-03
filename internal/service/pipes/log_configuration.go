// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pipes

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pipes/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func logConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"level": {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[types.LogLevel](),
				},
				"cloudwatch_logs_log_destination": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"log_group_arn": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"firehose_log_destination": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"delivery_stream_arn": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"s3_log_destination": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrBucketName: {
								Type:     schema.TypeString,
								Required: true,
							},
							"bucket_owner": {
								Type:     schema.TypeString,
								Required: true,
							},
							"output_format": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.S3OutputFormat](),
							},
							names.AttrPrefix: {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
			},
		},
	}
}

func expandPipeLogConfigurationParameters(tfMap map[string]interface{}) *types.PipeLogConfigurationParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeLogConfigurationParameters{}

	if v, ok := tfMap["level"].(string); ok && v != "" {
		apiObject.Level = types.LogLevel(v)
	}

	if v, ok := tfMap["cloudwatch_logs_log_destination"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.CloudwatchLogsLogDestination = expandCloudWatchLogsLogDestinationParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["firehose_log_destination"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.FirehoseLogDestination = expandFirehoseLogDestinationParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["s3_log_destination"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.S3LogDestination = expandS3LogDestinationParameters(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCloudWatchLogsLogDestinationParameters(tfMap map[string]interface{}) *types.CloudwatchLogsLogDestinationParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.CloudwatchLogsLogDestinationParameters{}

	if v, ok := tfMap["log_group_arn"].(string); ok && v != "" {
		apiObject.LogGroupArn = aws.String(v)
	}

	return apiObject
}

func expandFirehoseLogDestinationParameters(tfMap map[string]interface{}) *types.FirehoseLogDestinationParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.FirehoseLogDestinationParameters{}

	if v, ok := tfMap["delivery_stream_arn"].(string); ok && v != "" {
		apiObject.DeliveryStreamArn = aws.String(v)
	}

	return apiObject
}

func expandS3LogDestinationParameters(tfMap map[string]interface{}) *types.S3LogDestinationParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.S3LogDestinationParameters{}

	if v, ok := tfMap[names.AttrBucketName].(string); ok && v != "" {
		apiObject.BucketName = aws.String(v)
	}

	if v, ok := tfMap["bucket_owner"].(string); ok && v != "" {
		apiObject.BucketOwner = aws.String(v)
	}

	if v, ok := tfMap["output_format"].(string); ok && v != "" {
		apiObject.OutputFormat = types.S3OutputFormat(v)
	}

	if v, ok := tfMap[names.AttrPrefix].(string); ok && v != "" {
		apiObject.Prefix = aws.String(v)
	}

	return apiObject
}

func flattenPipeLogConfiguration(apiObject *types.PipeLogConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Level; v != "" {
		tfMap["level"] = v
	}

	if v := apiObject.CloudwatchLogsLogDestination; v != nil {
		tfMap["cloudwatch_logs_log_destination"] = []interface{}{flattenCloudWatchLogsLogDestination(v)}
	}

	if v := apiObject.FirehoseLogDestination; v != nil {
		tfMap["firehose_log_destination"] = []interface{}{flattenFirehoseLogDestination(v)}
	}

	if v := apiObject.S3LogDestination; v != nil {
		tfMap["s3_log_destination"] = []interface{}{flattenS3LogDestination(v)}
	}

	return tfMap
}

func flattenCloudWatchLogsLogDestination(apiObject *types.CloudwatchLogsLogDestination) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LogGroupArn; v != nil {
		tfMap["log_group_arn"] = aws.ToString(v)
	}

	return tfMap
}

func flattenFirehoseLogDestination(apiObject *types.FirehoseLogDestination) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DeliveryStreamArn; v != nil {
		tfMap["delivery_stream_arn"] = aws.ToString(v)
	}

	return tfMap
}

func flattenS3LogDestination(apiObject *types.S3LogDestination) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BucketName; v != nil {
		tfMap[names.AttrBucketName] = aws.ToString(v)
	}

	if v := apiObject.BucketOwner; v != nil {
		tfMap["bucket_owner"] = aws.ToString(v)
	}

	if v := apiObject.OutputFormat; v != "" {
		tfMap["output_format"] = v
	}

	if v := apiObject.Prefix; v != nil {
		tfMap[names.AttrPrefix] = aws.ToString(v)
	}

	return tfMap
}
