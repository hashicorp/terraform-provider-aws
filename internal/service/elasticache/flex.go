package elasticache

import (
	"github.com/aws/aws-sdk-go/service/elasticache"
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

func flattenSecurityGroupNames(securityGroups []*elasticache.CacheSecurityGroupMembership) []string {
	result := make([]string, 0, len(securityGroups))
	for _, sg := range securityGroups {
		if sg.CacheSecurityGroupName != nil {
			result = append(result, *sg.CacheSecurityGroupName)
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
