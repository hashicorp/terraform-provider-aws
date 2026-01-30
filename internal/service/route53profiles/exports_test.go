// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53profiles

var (
	Route53Profile                   = newProfileResource
	FindProfileByID                  = findProfileByID
	Route53ProfileAssocation         = newAssociationResource
	FindAssociationByID              = findAssociationByID
	Route53ProfileResourceAssocation = newResourceAssociationResource
	FindResourceAssociationByID      = findResourceAssociationByID
)
