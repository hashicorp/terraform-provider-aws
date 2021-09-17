package cloudtrail

const (
	ResourceTypeDynamoDBTable  = "AWS::DynamoDB::Table"
	ResourceTypeLambdaFunction = "AWS::Lambda::Function"
	ResourceTypeS3Object       = "AWS::S3::Object"
)

func ResourceType_Values() []string {
	return []string{
		ResourceTypeDynamoDBTable,
		ResourceTypeLambdaFunction,
		ResourceTypeS3Object,
	}
}

const (
	FieldEventCategory = "eventCategory"
	FieldEventName     = "eventName"
	FieldEventSource   = "eventSource"
	FieldReadOnly      = "readOnly"
	FieldResourcesARN  = "resources.ARN"
	FieldResourcesType = "resources.type"
)

func Field_Values() []string {
	return []string{
		FieldEventCategory,
		FieldEventName,
		FieldEventSource,
		FieldReadOnly,
		FieldResourcesARN,
		FieldResourcesType,
	}
}
