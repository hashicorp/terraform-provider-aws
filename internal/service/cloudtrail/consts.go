// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudtrail

import (
	"time"
)

const (
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
	fieldErrorCode                    = "errorCode"
	fieldEventCategory                = "eventCategory"
	fieldEventName                    = "eventName"
	fieldEventSource                  = "eventSource"
	fieldEventType                    = "eventType"
	fieldReadOnly                     = "readOnly"
	fieldResourcesARN                 = "resources.ARN"
	fieldResourcesType                = "resources.type"
	fieldSessionCredentialFromConsole = "sessionCredentialFromConsole"
	fieldUserIdentityARN              = "userIdentity.arn"
	fieldVPCEndpointID                = "vpcEndpointId"
)

func field_Values() []string {
	return []string{
		fieldErrorCode,
		fieldEventCategory,
		fieldEventName,
		fieldEventSource,
		fieldEventType,
		fieldReadOnly,
		fieldResourcesARN,
		fieldResourcesType,
		fieldSessionCredentialFromConsole,
		fieldUserIdentityARN,
		fieldVPCEndpointID,
	}
}

const (
	propagationTimeout = 2 * time.Minute
)

const (
	servicePrincipal = "cloudtrail.amazonaws.com"
)
