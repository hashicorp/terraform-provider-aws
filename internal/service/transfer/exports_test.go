// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package transfer

// Exports for use in tests only.
var (
	ResourceAccess              = resourceAccess
	ResourceAgreement           = resourceAgreement
	ResourceCertificate         = resourceCertificate
	ResourceConnector           = resourceConnector
	ResourceHostKey             = newHostKeyResource
	ResourceProfile             = resourceProfile
	ResourceServer              = resourceServer
	ResourceSSHKey              = resourceSSHKey
	ResourceTag                 = resourceTag
	ResourceUser                = resourceUser
	ResourceWebApp              = newWebAppResource
	ResourceWebAppCustomization = newWebAppCustomizationResource
	ResourceWorkflow            = resourceWorkflow

	FindAccessByTwoPartKey       = findAccessByTwoPartKey
	FindAgreementByTwoPartKey    = findAgreementByTwoPartKey
	FindCertificateByID          = findCertificateByID
	FindConnectorByID            = findConnectorByID
	FindHostKeyByTwoPartKey      = findHostKeyByTwoPartKey
	FindProfileByID              = findProfileByID
	FindServerByID               = findServerByID
	FindTag                      = findTag
	FindUserByTwoPartKey         = findUserByTwoPartKey
	FindUserSSHKeyByThreePartKey = findUserSSHKeyByThreePartKey
	FindWorkflowByID             = findWorkflowByID
	FindWebAppByID               = findWebAppByID
	FindWebAppCustomizationByID  = findWebAppCustomizationByID
)
