// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func flattenSecurityGroupIDs(securityGroups []*elasticache.SecurityGroupMembership) []string {
	result := make([]string, 0, len(securityGroups))
	for _, sg := range securityGroups {
		if sg.SecurityGroupId != nil {
			result = append(result, *sg.SecurityGroupId)
		}
	}
	return result
}

func flattenLogDeliveryConfigurations(logDeliveryConfiguration []*elasticache.LogDeliveryConfiguration) []map[string]interface{} {
	if len(logDeliveryConfiguration) == 0 {
		return nil
	}

	var logDeliveryConfigurations []map[string]interface{}
	for _, v := range logDeliveryConfiguration {
		logDeliveryConfig := make(map[string]interface{})

		switch aws.StringValue(v.DestinationType) {
		case elasticache.DestinationTypeKinesisFirehose:
			logDeliveryConfig["destination"] = aws.StringValue(v.DestinationDetails.KinesisFirehoseDetails.DeliveryStream)
		case elasticache.DestinationTypeCloudwatchLogs:
			logDeliveryConfig["destination"] = aws.StringValue(v.DestinationDetails.CloudWatchLogsDetails.LogGroup)
		}

		logDeliveryConfig["destination_type"] = aws.StringValue(v.DestinationType)
		logDeliveryConfig["log_format"] = aws.StringValue(v.LogFormat)
		logDeliveryConfig["log_type"] = aws.StringValue(v.LogType)
		logDeliveryConfigurations = append(logDeliveryConfigurations, logDeliveryConfig)
	}

	return logDeliveryConfigurations
}

func expandEmptyLogDeliveryConfigurations(v map[string]interface{}) elasticache.LogDeliveryConfigurationRequest {
	logDeliveryConfigurationRequest := elasticache.LogDeliveryConfigurationRequest{}
	logDeliveryConfigurationRequest.SetEnabled(false)
	logDeliveryConfigurationRequest.SetLogType(v["log_type"].(string))

	return logDeliveryConfigurationRequest
}

func expandLogDeliveryConfigurations(v map[string]interface{}) elasticache.LogDeliveryConfigurationRequest {
	logDeliveryConfigurationRequest := elasticache.LogDeliveryConfigurationRequest{}

	logDeliveryConfigurationRequest.LogType = aws.String(v["log_type"].(string))
	logDeliveryConfigurationRequest.DestinationType = aws.String(v["destination_type"].(string))
	logDeliveryConfigurationRequest.LogFormat = aws.String(v["log_format"].(string))
	destinationDetails := elasticache.DestinationDetails{}

	switch v["destination_type"].(string) {
	case elasticache.DestinationTypeCloudwatchLogs:
		destinationDetails.CloudWatchLogsDetails = &elasticache.CloudWatchLogsDestinationDetails{
			LogGroup: aws.String(v["destination"].(string)),
		}
	case elasticache.DestinationTypeKinesisFirehose:
		destinationDetails.KinesisFirehoseDetails = &elasticache.KinesisFirehoseDestinationDetails{
			DeliveryStream: aws.String(v["destination"].(string)),
		}
	}

	logDeliveryConfigurationRequest.DestinationDetails = &destinationDetails

	return logDeliveryConfigurationRequest
}

func expandNodeGroupConfiguration(v map[string]interface{}) elasticache.NodeGroupConfiguration {
	nodeGroupConfiguration := elasticache.NodeGroupConfiguration{}

	if v, ok := v["node_group_id"].(string); ok && v != "" {
		nodeGroupConfiguration.NodeGroupId = aws.String(v)
	}
	if v, ok := v["primary_availability_zone"].(string); ok && v != "" {
		nodeGroupConfiguration.PrimaryAvailabilityZone = aws.String(v)
	}
	if v, ok := v["primary_outpost_arn"].(string); ok && v != "" {
		nodeGroupConfiguration.PrimaryOutpostArn = aws.String(v)
	}
	if v, ok := v["primary_outpost_arn"].(string); ok && v != "" {
		nodeGroupConfiguration.PrimaryOutpostArn = aws.String(v)
	}
	if v, ok := v["replica_count"].(int64); ok && v != 0 {
		nodeGroupConfiguration.ReplicaCount = aws.Int64(v)
	}
	if raz := v["replica_availability_zones"].(*schema.Set); raz.Len() > 0 {
		nodeGroupConfiguration.ReplicaAvailabilityZones = flex.ExpandStringSet(raz)
	}
	if roa := v["replica_outpost_arns"].(*schema.Set); roa.Len() > 0 {
		nodeGroupConfiguration.ReplicaOutpostArns = flex.ExpandStringSet(roa)
	}
	if v, ok := v["replica_count"].(int64); ok && v != 0 {
		nodeGroupConfiguration.ReplicaCount = aws.Int64(v)
	}
	if v, ok := v["slots"].(string); ok && v != "" {
		nodeGroupConfiguration.Slots = aws.String(v)
	}

	return nodeGroupConfiguration
}
