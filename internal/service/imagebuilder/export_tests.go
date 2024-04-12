// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

// Exports for use in tests only.
var (
	ResourceLifecyclePolicy = newResourceLifecyclePolicy

	FindLifecyclePolicyByARN = findLifecyclePolicyByARN
)
