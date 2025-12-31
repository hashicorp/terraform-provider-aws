// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package codeconnections

// Exports for use in tests only.
var (
	ResourceConnection = newConnectionResource
	ResourceHost       = newHostResource

	FindConnectionByARN = findConnectionByARN
	FindHostByARN       = findHostByARN
)
