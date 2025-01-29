// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package worklink

// Exports for use in tests only.
var (
	FindFleetByARN                            = findFleetByARN
	FindWebsiteCertificateAuthorityByARNAndID = findWebsiteCertificateAuthorityByARNAndID

	DecodeWebsiteCertificateAuthorityAssociationResourceID = decodeWebsiteCertificateAuthorityAssociationResourceID
	FleetStateRefresh                                      = fleetStateRefresh
	WebsiteCertificateAuthorityAssociationStateRefresh     = websiteCertificateAuthorityAssociationStateRefresh
)
