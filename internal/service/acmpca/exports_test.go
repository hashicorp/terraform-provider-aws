// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package acmpca

// Exports for use in tests only.
var (
	ResourceCertificate                     = resourceCertificate
	ResourceCertificateAuthority            = resourceCertificateAuthority
	ResourceCertificateAuthorityCertificate = resourceCertificateAuthorityCertificate
	ResourcePermission                      = resourcePermission
	ResourcePolicy                          = resourcePolicy

	FindCertificateAuthorityCertificateByARN = findCertificateAuthorityCertificateByARN
	FindPermissionByThreePartKey             = findPermissionByThreePartKey
	FindPolicyByARN                          = findPolicyByARN
	ValidTemplateARN                         = validTemplateARN
)
