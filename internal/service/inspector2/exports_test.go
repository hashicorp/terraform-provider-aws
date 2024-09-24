// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector2

// Exports for use in tests only.
var (
	ResourceDelegatedAdminAccount = resourceDelegatedAdminAccount

	FindDelegatedAdminAccountByID     = findDelegatedAdminAccountByID
	WaitDelegatedAdminAccountDisabled = waitDelegatedAdminAccountDisabled
	WaitDelegatedAdminAccountEnabled  = waitDelegatedAdminAccountEnabled

	EnablerID      = enablerID
	ParseEnablerID = parseEnablerID
)
