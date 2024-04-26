// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pipes

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pipes/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
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
					ValidateDiagFunc: enum.Validate[awstypes.LogLevel](),
				},
				"log_group_arn": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"delivery_stream_arn": {
					Type:     schema.TypeString,
					Optional: true,
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
		apiObject.Level = aws.String(v)
	}

	if v, ok := tfMap["log_group_arn"].(string); ok && v != "" {
		apiObject.CloudwatchLogsLogDestination = &types.CloudwatchLogsLogDestinationParameters{
			LogGroupArn: aws.String(v),
		}
	}

	if v, ok := tfMap["delivery_stream_arn"].(string); ok && v != "" {
		apiObject.FirehoseLogDestination = &types.FirehoseLogDestinationParameters{
			DeliveryStreamArn: aws.String(v),
		}
	}

	return apiObject
}

func flattenPipeLogConfigurationParameters(apiObject *types.PipeLogConfigurationParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Level; v != nil {
		tfMap["level"] = aws.ToString(v)
	}

	if v := apiObject.CloudwatchLogsLogDestination; v != nil {
		tfMap["log_group_arn"] = aws.ToString(apiObject.CloudwatchLogsLogDestination.LogGroupArn)
	}

	if v := apiObject.FirehoseLogDestination; v != nil {
		tfMap["delivery_stream_arn"] = aws.ToString(apiObject.FirehoseLogDestination.DeliveryStreamArn)
	}

	return tfMap
}
