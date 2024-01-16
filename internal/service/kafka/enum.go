// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

const (
	clusterOperationStatePending          = "PENDING"
	clusterOperationStateUpdateComplete   = "UPDATE_COMPLETE"
	clusterOperationStateUpdateFailed     = "UPDATE_FAILED"
	clusterOperationStateUpdateInProgress = "UPDATE_IN_PROGRESS"
)

type publicAccessType string

const (
	publicAccessTypeDisabled            publicAccessType = "DISABLED"
	publicAccessTypeServiceProvidedEIPs publicAccessType = "SERVICE_PROVIDED_EIPS"
)

func (publicAccessType) Values() []publicAccessType {
	return []publicAccessType{
		publicAccessTypeDisabled,
		publicAccessTypeServiceProvidedEIPs,
	}
}
