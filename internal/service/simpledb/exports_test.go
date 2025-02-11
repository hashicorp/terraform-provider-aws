// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package simpledb

// Exports for use in tests only.
var (
	ResourceDomain = newDomainResource

	FindDomainByName = findDomainByName
	SimpleDBConn     = simpleDBConn // nosemgrep:ci.simpledb-in-var-name
)
