// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions

// Exports for use in tests only.
var (
	ResourceIdentitySource = newResourceIdentitySource
	ResourcePolicy         = newResourcePolicy
	ResourcePolicyStore    = newResourcePolicyStore
	ResourcePolicyTemplate = newResourcePolicyTemplate
	ResourceSchema         = newResourceSchema

	FindIdentitySourceByIDAndPolicyStoreID = findIdentitySourceByIDAndPolicyStoreID
	FindPolicyByID                         = findPolicyByID
	FindPolicyStoreByID                    = findPolicyStoreByID
	FindPolicyTemplateByID                 = findPolicyTemplateByID
	FindSchemaByPolicyStoreID              = findSchemaByPolicyStoreID
)

var (
	PolicyTemplateParseID = policyTemplateParseID
)
