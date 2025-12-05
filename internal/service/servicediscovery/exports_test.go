// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package servicediscovery

// Exports for use in tests only.
var (
	ResourceHTTPNamespace       = resourceHTTPNamespace
	ResourceInstance            = resourceInstance
	ResourcePrivateDNSNamespace = resourcePrivateDNSNamespace
	ResourcePublicDNSNamespace  = resourcePublicDNSNamespace
	ResourceService             = resourceService

	FindInstanceByTwoPartKey = findInstanceByTwoPartKey
	FindNamespaceByID        = findNamespaceByID
	FindServiceByID          = findServiceByID
	ValidNamespaceName       = validNamespaceName
)
