// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

// Exports for use in tests only.
var (
	ResourceAccess = resourceAccess
	ResourceServer = resourceServer
	ResourceTag    = resourceTag

	FindAccessByTwoPartKey    = findAccessByTwoPartKey
	FindAgreementByTwoPartKey = findAgreementByTwoPartKey
	FindCertificateByID       = findCertificateByID
	FindConnectorByID         = findConnectorByID
	FindProfileByID           = findProfileByID
	FindServerByID            = findServerByID
	FindWorkflowByID          = findWorkflowByID
)
