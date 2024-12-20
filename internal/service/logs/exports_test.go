// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

// Exports for use in tests only.
var (
	ResourceAccountPolicy        = resourceAccountPolicy
	ResourceAnomalyDetector      = newAnomalyDetectorResource
	ResourceDataProtectionPolicy = resourceDataProtectionPolicy
	ResourceDestination          = resourceDestination
	ResourceDestinationPolicy    = resourceDestinationPolicy
	ResourceGroup                = resourceGroup
	ResourceIndexPolicy          = newIndexPolicyResource
	ResourceMetricFilter         = resourceMetricFilter
	ResourceQueryDefinition      = resourceQueryDefinition
	ResourceResourcePolicy       = resourceResourcePolicy
	ResourceStream               = resourceStream
	ResourceSubscriptionFilter   = resourceSubscriptionFilter

	FindAccountPolicyByTwoPartKey          = findAccountPolicyByTwoPartKey
	FindDataProtectionPolicyByLogGroupName = findDataProtectionPolicyByLogGroupName
	FindDestinationByName                  = findDestinationByName
	FindDestinationPolicyByName            = findDestinationPolicyByName
	FindIndexPolicyByLogGroupName          = findIndexPolicyByLogGroupName
	FindLogAnomalyDetectorByARN            = findLogAnomalyDetectorByARN
	FindLogGroupByName                     = findLogGroupByName
	FindLogStreamByTwoPartKey              = findLogStreamByTwoPartKey // nosemgrep:ci.logs-in-var-name
	FindMetricFilterByTwoPartKey           = findMetricFilterByTwoPartKey
	FindQueryDefinitionByTwoPartKey        = findQueryDefinitionByTwoPartKey
	FindResourcePolicyByName               = findResourcePolicyByName
	FindSubscriptionFilterByTwoPartKey     = findSubscriptionFilterByTwoPartKey

	TrimLogGroupARNWildcardSuffix          = trimLogGroupARNWildcardSuffix
	ValidLogGroupName                      = validLogGroupName
	ValidLogGroupNamePrefix                = validLogGroupNamePrefix
	ValidLogMetricFilterName               = validLogMetricFilterName
	ValidLogMetricFilterTransformationName = validLogMetricFilterTransformationName
	ValidLogStreamName                     = validLogStreamName // nosemgrep:ci.logs-in-var-name
)
