// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fms

// Exports for use in tests only.
var (
	ResourceAdminAccount = resourceAdminAccount
	ResourcePolicy       = resourcePolicy
	ResourceSet          = newResourceSetResource

	FindAdminAccount          = findAdminAccount
	FindPolicyByID            = findPolicyByID
	FindResourceSetByID       = findResourceSetByID
	RemoveEmptyFieldsFromJSON = removeEmptyFieldsFromJSON
)
