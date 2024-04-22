// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

const (
	ResNameDistribution = "Distribution"
)

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
	keyValueStoreStatusProvisioning = "PROVISIONING"
	keyValueStoreStatusReady        = "READY"
)
