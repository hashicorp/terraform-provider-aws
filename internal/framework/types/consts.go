// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

// ProviderErrorDetailPrefix contains instructions for reporting provider errors to provider developers
const ProviderErrorDetailPrefix = "An unexpected error was encountered trying to validate an attribute value. " +
	"This is always an error in the provider. Please report the following to the provider developer:\n\n"
