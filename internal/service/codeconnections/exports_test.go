// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeconnections

// Exports for use in tests only.
var (
	ResourceConnection = newConnectionResource
	ResourceHost       = newHostResource

	FindConnectionByARN = findConnectionByARN
	FindHostByARN       = findHostByARN
)
