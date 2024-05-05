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
