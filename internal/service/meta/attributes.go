// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package meta

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// idAttributeDeprecatedWithAlternate is a variant of datasourceattribute.IDAttributeDeprecatedWithAlternate
// that allows it to be set, preventing a breaking change.
func idAttributeDeprecatedWithAlternate(altPath path.Path) schema.StringAttribute {
	return schema.StringAttribute{
		Optional:           true,
		Computed:           true,
		DeprecationMessage: deprecatedWithAlternateMessage(altPath),
	}
}

// idAttributeDeprecatedNoReplacement is a variant of datasourceattribute.IDAttributeDeprecatedNoReplacement
// that allows it to be set, preventing a breaking change.
func idAttributeDeprecatedNoReplacement() schema.StringAttribute {
	return schema.StringAttribute{
		Optional:           true,
		Computed:           true,
		DeprecationMessage: "This attribute will be removed in a future version of the provider.",
	}
}

func deprecatedWithAlternateMessage(altPath path.Path) string {
	return fmt.Sprintf("Use '%s' instead. This attribute will be removed in a future version of the provider.", altPath.String())
}
