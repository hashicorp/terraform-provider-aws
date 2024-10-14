// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53profiles

var (
	Route53Profile                   = newResourceProfile
	FindProfileByID                  = findProfileByID
	Route53ProfileAssocation         = newResourceAssociation
	FindAssociationByID              = findAssociationByID
	Route53ProfileResourceAssocation = newResourceResourceAssociation
	FindResourceAssociationByID      = findResourceAssociationByID
)
