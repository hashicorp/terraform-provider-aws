// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

// Exports for use in tests only.
var (
	ResourceAccess      = resourceAccess
	ResourceAgreement   = resourceAgreement
	ResourceCertificate = resourceCertificate
	ResourceConnector   = resourceConnector
	ResourceProfile     = resourceProfile
	ResourceServer      = resourceServer
	ResourceSSHKey      = resourceSSHKey
	ResourceTag         = resourceTag
	ResourceUser        = resourceUser

	FindAccessByTwoPartKey       = findAccessByTwoPartKey
	FindAgreementByTwoPartKey    = findAgreementByTwoPartKey
	FindCertificateByID          = findCertificateByID
	FindConnectorByID            = findConnectorByID
	FindProfileByID              = findProfileByID
	FindServerByID               = findServerByID
	FindUserByTwoPartKey         = findUserByTwoPartKey
	FindUserSSHKeyByThreePartKey = findUserSSHKeyByThreePartKey
	FindWorkflowByID             = findWorkflowByID
)
