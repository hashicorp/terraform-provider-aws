// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func flattenSecurityGroupIDs(apiObjects []awstypes.SecurityGroupMembership) []string {
	return tfslices.ApplyToAll(apiObjects, func(v awstypes.SecurityGroupMembership) string {
		return aws.ToString(v.SecurityGroupId)
	})
}

func flattenLogDeliveryConfigurations(apiObjects []awstypes.LogDeliveryConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]interface{})

		switch apiObject.DestinationType {
		case awstypes.DestinationTypeKinesisFirehose:
			tfMap[names.AttrDestination] = aws.ToString(apiObject.DestinationDetails.KinesisFirehoseDetails.DeliveryStream)
		case awstypes.DestinationTypeCloudWatchLogs:
			tfMap[names.AttrDestination] = aws.ToString(apiObject.DestinationDetails.CloudWatchLogsDetails.LogGroup)
		}
		tfMap["destination_type"] = apiObject.DestinationType
		tfMap["log_format"] = apiObject.LogFormat
		tfMap["log_type"] = apiObject.LogType

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandEmptyLogDeliveryConfigurationRequest(tfMap map[string]interface{}) awstypes.LogDeliveryConfigurationRequest {
	apiObject := awstypes.LogDeliveryConfigurationRequest{}

	apiObject.Enabled = aws.Bool(false)
	apiObject.LogType = awstypes.LogType(tfMap["log_type"].(string))

	return apiObject
}

func expandLogDeliveryConfigurationRequests(v map[string]interface{}) awstypes.LogDeliveryConfigurationRequest {
	apiObject := awstypes.LogDeliveryConfigurationRequest{}

	destinationType := awstypes.DestinationType(v["destination_type"].(string))
	apiObject.DestinationType = destinationType
	destinationDetails := &awstypes.DestinationDetails{}
	switch destinationType {
	case awstypes.DestinationTypeCloudWatchLogs:
		destinationDetails.CloudWatchLogsDetails = &awstypes.CloudWatchLogsDestinationDetails{
			LogGroup: aws.String(v[names.AttrDestination].(string)),
		}
	case awstypes.DestinationTypeKinesisFirehose:
		destinationDetails.KinesisFirehoseDetails = &awstypes.KinesisFirehoseDestinationDetails{
			DeliveryStream: aws.String(v[names.AttrDestination].(string)),
		}
	}
	apiObject.DestinationDetails = destinationDetails
	apiObject.LogType = awstypes.LogType(v["log_type"].(string))
	apiObject.LogFormat = awstypes.LogFormat(v["log_format"].(string))

	return apiObject
}
