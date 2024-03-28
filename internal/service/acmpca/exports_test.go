// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acmpca

// Exports for use in tests only.
var (
	ResourceCertificate                     = resourceCertificate
	ResourceCertificateAuthority            = resourceCertificateAuthority
	ResourceCertificateAuthorityCertificate = resourceCertificateAuthorityCertificate
	ResourcePermission                      = resourcePermission

	FindCertificateAuthorityCertificateByARN = findCertificateAuthorityCertificateByARN
	FindCertificateByTwoPartKey              = findCertificateByTwoPartKey
	FindPermissionByThreePartKey             = findPermissionByThreePartKey
	ValidTemplateARN                         = validTemplateARN
)
