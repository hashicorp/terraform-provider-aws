// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

// Exports for use in tests only.
var (
	ResourceAssertion       = newResourceAssertion
	ResourceInputSource     = newResourceInputSource
	ResourcePolicy          = newResourcePolicy
	ResourceService         = newResourceService
	ResourceServiceFunction = newResourceServiceFunction
	ResourceSystem          = newResourceSystem

	FindAssertionByID       = findAssertionByID
	FindInputSourceByID     = findInputSourceByID
	FindPolicyByARN         = findPolicyByARN
	FindServiceByARN        = findServiceByARN
	FindServiceFunctionByID = findServiceFunctionByID
	FindSystemByARN         = findSystemByARN
)
