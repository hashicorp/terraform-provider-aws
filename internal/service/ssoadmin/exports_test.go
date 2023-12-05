// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

// Exports for use in tests only.
var (
	ResourceApplication           = newResourceApplication
	ResourceApplicationAssignment = newResourceApplicationAssignment

	FindApplicationByID           = findApplicationByID
	FindApplicationAssignmentByID = findApplicationAssignmentByID
)
