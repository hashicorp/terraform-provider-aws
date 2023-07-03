// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

const (
	ClusterOperationStatePending          = "PENDING"
	ClusterOperationStateUpdateComplete   = "UPDATE_COMPLETE"
	ClusterOperationStateUpdateFailed     = "UPDATE_FAILED"
	ClusterOperationStateUpdateInProgress = "UPDATE_IN_PROGRESS"
)

const (
	PublicAccessTypeDisabled            = "DISABLED"
	PublicAccessTypeServiceProvidedEIPs = "SERVICE_PROVIDED_EIPS"
)

func PublicAccessType_Values() []string {
	return []string{
		PublicAccessTypeDisabled,
		PublicAccessTypeServiceProvidedEIPs,
	}
}
