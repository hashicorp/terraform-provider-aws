// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicecatalogappregistry

// Exports for use in tests only.
var (
	FindApplicationByID                       = findApplicationByID
	FindAttributeGroupByID                    = findAttributeGroupByID
	FindAttributeGroupAssociationByTwoPartKey = findAttributeGroupAssociationByTwoPartKey

	ResourceApplication               = newApplicationResource
	ResourceAttributeGroup            = newAttributeGroupResource
	ResourceAttributeGroupAssociation = newAttributeGroupAssociationResource
)
