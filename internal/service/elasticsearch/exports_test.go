// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package elasticsearch

// Exports for use in tests only.
var (
	ResourceDomain            = resourceDomain
	ResourceDomainPolicy      = resourceDomainPolicy
	ResourceDomainSAMLOptions = resourceDomainSAMLOptions
	ResourceVPCEndpoint       = resourceVPCEndpoint

	FindDomainByName                 = findDomainByName
	FindDomainSAMLOptionByDomainName = findDomainSAMLOptionByDomainName
	FindVPCEndpointByID              = findVPCEndpointByID
	VPCEndpointsError                = vpcEndpointsError
	WaitDomainCreated                = waitDomainCreated
)
