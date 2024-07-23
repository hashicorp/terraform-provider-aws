// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package paymentcryptography

// Exports for use in tests only.
var (
	ResourceKey      = newResourceKey
	ResourceKeyAlias = newResourceKeyAlias

	FindKeyByID        = findKeyByID
	FindKeyAliasByName = findkeyAliasByName
)
