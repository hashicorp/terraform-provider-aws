// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package acm

// Exports for use in tests only.
var (
	FindACMEEndpointByARN          = findACMEEndpointByARN
	FindCertificateByARN           = findCertificateByARN
	FindCertificateValidationByARN = findCertificateValidationByARN
	WaitCertificateRenewed         = waitCertificateRenewed

	ResourceACMEEndpoint          = newACMEEndpointResource
	ResourceCertificate           = resourceCertificate
	ResourceCertificateValidation = resourceCertificateValidation
)
