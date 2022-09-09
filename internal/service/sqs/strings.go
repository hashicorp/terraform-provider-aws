package sqs

import "github.com/aws/aws-sdk-go/service/sqs"

func getSchemaKey(attributeName string) string {
	switch attributeName {
	case sqs.QueueAttributeNamePolicy:
		return "policy"
	case sqs.QueueAttributeNameRedrivePolicy:
		return "redrive_policy"
	default:
		return ""
	}
}
