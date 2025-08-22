// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions

// Exports for use in tests only.
var (
	ResourceIdentitySource = newIdentitySourceResource
	ResourcePolicy         = newPolicyResource
	ResourcePolicyStore    = newPolicyStoreResource
	ResourcePolicyTemplate = newPolicyTemplateResource
	ResourceSchema         = newSchemaResource

	FindIdentitySourceByIDAndPolicyStoreID = findIdentitySourceByIDAndPolicyStoreID
	FindPolicyByID                         = findPolicyByID
	FindPolicyStoreByID                    = findPolicyStoreByID
	FindPolicyTemplateByID                 = findPolicyTemplateByID
	FindSchemaByPolicyStoreID              = findSchemaByPolicyStoreID
)

var (
	PolicyTemplateParseID = policyTemplateParseID
)
