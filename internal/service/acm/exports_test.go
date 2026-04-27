// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package acm

// Exports for use in tests only.
var (
	FindCertificateByARN           = findCertificateByARN
	FindCertificateValidationByARN = findCertificateValidationByARN
	WaitCertificateRenewed         = waitCertificateRenewed

	ResourceCertificate           = resourceCertificate
	ResourceCertificateValidation = resourceCertificateValidation
)
