// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions

// Exports for use in tests only.
var (
	ResourcePolicyStore = newResourcePolicyStore

	FindPolicyStoreByID = findPolicyStoreByID
)
