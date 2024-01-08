// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions

// Exports for use in tests only.
var (
	ResourcePolicyStore    = newResourcePolicyStore
	ResourcePolicyTemplate = newResourcePolicyTemplate
	ResourceSchema         = newResourceSchema

	FindPolicyStoreByID       = findPolicyStoreByID
	FindPolicyTemplateByID    = findPolicyTemplateByID
	FindSchemaByPolicyStoreID = findSchemaByPolicyStoreID
)

var (
	PolicyTemplateParseID = policyTemplateParseID
)
