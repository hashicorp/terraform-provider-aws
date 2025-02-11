// Copyright (c) HashiCorp, Inc.
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
