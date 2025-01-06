// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

const (
	jobDefinitionStatusActive   string = "ACTIVE"
	jobDefinitionStatusInactive string = "INACTIVE"
)

func jobDefinitionStatus_Values() []string {
	return []string{
		jobDefinitionStatusInactive,
		jobDefinitionStatusActive,
	}
}

const (
	imagePullPolicyAlways       = "Always"
	imagePullPolicyIfNotPresent = "IfNotPresent"
	imagePullPolicyNever        = "Never"
)

func imagePullPolicy_Values() []string {
	return []string{
		imagePullPolicyAlways,
		imagePullPolicyIfNotPresent,
		imagePullPolicyNever,
	}
}

const (
	dnsPolicyDefault                 = "Default"
	dnsPolicyClusterFirst            = "ClusterFirst"
	dnsPolicyClusterFirstWithHostNet = "ClusterFirstWithHostNet"
)

func dnsPolicy_Values() []string {
	return []string{
		dnsPolicyDefault,
		dnsPolicyClusterFirst,
		dnsPolicyClusterFirstWithHostNet,
	}
}
