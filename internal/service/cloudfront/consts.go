// Copyright (c) HashiCorp, Inc.
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
	distributionStatusDeployed   = "Deployed"
	distributionStatusInProgress = "InProgress"
)

const (
	keyValueStoreStatusProvisioning = "PROVISIONING"
	keyValueStoreStatusReady        = "READY"
)
