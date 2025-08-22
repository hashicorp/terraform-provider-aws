// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qldb

// Exports for use in tests only.
var (
	FindLedgerByName       = findLedgerByName
	FindStreamByTwoPartKey = findStreamByTwoPartKey

	ResourceLedger = resourceLedger
	ResourceStream = resourceStream
)
