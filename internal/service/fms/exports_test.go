// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fms

// Exports for use in tests only.
var (
	ResourceAdminAccount = resourceAdminAccount
	ResourcePolicy       = resourcePolicy

	FindAdminAccount          = findAdminAccount
	FindPolicyByID            = findPolicyByID
	RemoveEmptyFieldsFromJSON = removeEmptyFieldsFromJSON
)
