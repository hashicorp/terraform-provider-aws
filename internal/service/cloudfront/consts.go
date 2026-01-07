// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront

type streamType string

const (
	streamTypeKinesis streamType = "Kinesis"
)

func (streamType) Values() []streamType {
	return []streamType{
		streamTypeKinesis,
	}
}

const (
	connectionFunctionStatusPublishing   = "PUBLISHING"
	connectionFunctionStatusUnassociated = "UNASSOCIATED"
	connectionFunctionStatusUnpublished  = "UNPUBLISHED"
)

const (
	distributionStatusDeployed   = "Deployed"
	distributionStatusInProgress = "InProgress"
)

const (
	connectionGroupStatusDeployed   = "Deployed"
	connectionGroupStatusInProgress = "InProgress"
)

const (
	keyValueStoreStatusProvisioning = "PROVISIONING"
	keyValueStoreStatusReady        = "READY"
)

const (
	vpcOriginStatusDeployed  = "Deployed"
	vpcOriginStatusDeploying = "Deploying"
)
