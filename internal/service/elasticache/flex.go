// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func flattenSecurityGroupIDs(securityGroups []awstypes.SecurityGroupMembership) []string {
	result := make([]string, 0, len(securityGroups))
	for _, sg := range securityGroups {
		if sg.SecurityGroupId != nil {
			result = append(result, *sg.SecurityGroupId)
		}
	}
	return result
}

func flattenLogDeliveryConfigurations(logDeliveryConfiguration []awstypes.LogDeliveryConfiguration) []map[string]interface{} {
	if len(logDeliveryConfiguration) == 0 {
		return nil
	}

	var logDeliveryConfigurations []map[string]interface{}
	for _, v := range logDeliveryConfiguration {
		logDeliveryConfig := make(map[string]interface{})

		switch v.DestinationType {
		case awstypes.DestinationTypeKinesisFirehose:
			logDeliveryConfig[names.AttrDestination] = aws.ToString(v.DestinationDetails.KinesisFirehoseDetails.DeliveryStream)
		case awstypes.DestinationTypeCloudWatchLogs:
			logDeliveryConfig[names.AttrDestination] = aws.ToString(v.DestinationDetails.CloudWatchLogsDetails.LogGroup)
		}

		logDeliveryConfig["destination_type"] = string(v.DestinationType)
		logDeliveryConfig["log_format"] = string(v.LogFormat)
		logDeliveryConfig["log_type"] = string(v.LogType)
		logDeliveryConfigurations = append(logDeliveryConfigurations, logDeliveryConfig)
	}

	return logDeliveryConfigurations
}

func expandEmptyLogDeliveryConfigurations(v map[string]interface{}) awstypes.LogDeliveryConfigurationRequest {
	logDeliveryConfigurationRequest := awstypes.LogDeliveryConfigurationRequest{}
	logDeliveryConfigurationRequest.Enabled = aws.Bool(false)
	logDeliveryConfigurationRequest.LogType = awstypes.LogType(v["log_type"].(string))

	return logDeliveryConfigurationRequest
}

func expandLogDeliveryConfigurations(v map[string]interface{}) awstypes.LogDeliveryConfigurationRequest {
	logDeliveryConfigurationRequest := awstypes.LogDeliveryConfigurationRequest{}

	logDeliveryConfigurationRequest.LogType = awstypes.LogType(v["log_type"].(string))
	logDeliveryConfigurationRequest.DestinationType = awstypes.DestinationType(v["destination_type"].(string))
	logDeliveryConfigurationRequest.LogFormat = awstypes.LogFormat(v["log_format"].(string))
	destinationDetails := awstypes.DestinationDetails{}

	switch v["destination_type"].(string) {
	case string(awstypes.DestinationTypeCloudWatchLogs):
		destinationDetails.CloudWatchLogsDetails = &awstypes.CloudWatchLogsDestinationDetails{
			LogGroup: aws.String(v[names.AttrDestination].(string)),
		}
	case string(awstypes.DestinationTypeKinesisFirehose):
		destinationDetails.KinesisFirehoseDetails = &awstypes.KinesisFirehoseDestinationDetails{
			DeliveryStream: aws.String(v[names.AttrDestination].(string)),
		}
	}

	logDeliveryConfigurationRequest.DestinationDetails = &destinationDetails

	return logDeliveryConfigurationRequest
}
