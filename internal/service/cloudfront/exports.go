// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

// Exports for use across service packages.
var (
	FindDistributionByID  = findDistributionByID
	ResourceKeyValueStore = newKeyValueStoreResource
)
