package cloudtrail

import "time"

const (
	ResCloudTrail = "Trail"

	resourceTypeDynamoDBTable  = "AWS::DynamoDB::Table"
	resourceTypeLambdaFunction = "AWS::Lambda::Function"
	resourceTypeS3Object       = "AWS::S3::Object"
)

func resourceType_Values() []string {
	return []string{
		resourceTypeDynamoDBTable,
		resourceTypeLambdaFunction,
		resourceTypeS3Object,
	}
}

const (
	fieldEventCategory = "eventCategory"
	fieldEventName     = "eventName"
	fieldEventSource   = "eventSource"
	fieldReadOnly      = "readOnly"
	fieldResourcesARN  = "resources.ARN"
	fieldResourcesType = "resources.type"
)

func field_Values() []string {
	return []string{
		fieldEventCategory,
		fieldEventName,
		fieldEventSource,
		fieldReadOnly,
		fieldResourcesARN,
		fieldResourcesType,
	}
}

const (
	propagationTimeout = 2 * time.Minute
)
