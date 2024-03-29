// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

const (
	StreamTypeKinesis = "Kinesis"

	ResNameDistribution         = "Distribution"
	ResNamePublicKey            = "Public Key"
	ResNameOriginAccessIdentity = "Origin Access Identity"
)

func StreamType_Values() []string {
	return []string{
		StreamTypeKinesis,
	}
}
