// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

const (
	JobDefinitionStatusInactive string = "INACTIVE"
	JobDefinitionStatusActive   string = "ACTIVE"
)

func JobDefinitionStatus_Values() []string {
	return []string{
		JobDefinitionStatusInactive,
		JobDefinitionStatusActive,
	}
}
