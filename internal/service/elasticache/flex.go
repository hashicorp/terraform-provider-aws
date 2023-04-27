package elasticache

import (
	"github.com/aws/aws-sdk-go/aws"
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
