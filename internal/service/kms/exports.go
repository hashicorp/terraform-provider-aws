// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

// Exports for use in other modules.
var (
	DiffSuppressKey             = diffSuppressKey
	DiffSuppressKeyOrAlias      = diffSuppressKeyOrAlias
	FindAliasByName             = findAliasByName
	FindDefaultKeyARNForService = findDefaultKeyARNForService
	FindKeyByID                 = findKeyByID // TODO Move to exports_test.go
	ValidateKey                 = validateKey
	ValidateKeyOrAlias          = validateKeyOrAlias
)
