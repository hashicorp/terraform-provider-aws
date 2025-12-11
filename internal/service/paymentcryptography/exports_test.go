// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package paymentcryptography

// Exports for use in tests only.
var (
	ResourceKey      = newKeyResource
	ResourceKeyAlias = newKeyAliasResource

	FindKeyByID        = findKeyByID
	FindKeyAliasByName = findkeyAliasByName
)
