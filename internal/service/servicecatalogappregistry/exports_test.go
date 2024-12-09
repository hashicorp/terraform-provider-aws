// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalogappregistry

// Exports for use in tests only.
var (
	FindApplicationByID               = findApplicationByID
	FindAttributeGroupByID            = findAttributeGroupByID
	ResourceApplication               = newResourceApplication
	ResourceAttributeGroup            = newResourceAttributeGroup
	ResourceAttributeGroupAssociation = newResourceAttributeGroupAssociation
)
