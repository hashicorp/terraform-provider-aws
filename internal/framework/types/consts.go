// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package types

import "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

// ProviderErrorDetailPrefix contains instructions for reporting provider errors to provider developers
const ProviderErrorDetailPrefix = "An unexpected error was encountered trying to validate an attribute value. " +
	"This is always an error in the provider. Please report the following to the provider developer:\n\n"

// CollectionLengthUnhandledAsZero is the standard options to use when calling Length on a
// framework collection type (List, Set, Map). It treats null and unknown values as zero-length.
var CollectionLengthUnhandledAsZero = basetypes.CollectionLengthOptions{
	UnhandledNullAsZero:    true,
	UnhandledUnknownAsZero: true,
}
