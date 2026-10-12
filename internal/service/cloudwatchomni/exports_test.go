// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

// Exports for use in tests only.
var (
	ResourceDomain = newDomainResource
	ResourceView   = newViewResource

	FindDomainByID = findDomainByID
	FindViewByName = findViewByName
)
