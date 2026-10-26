// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package datasourceattribute

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func IDAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Computed: true,
	}
}

func IDAttributeDeprecatedWithAlternate(altPath path.Path) schema.StringAttribute {
	return schema.StringAttribute{
		Computed:           true,
		DeprecationMessage: deprecatedWithAlternateMessage(altPath),
	}
}

func IDAttributeDeprecatedNoReplacement() schema.StringAttribute {
	return schema.StringAttribute{
		Computed:           true,
		DeprecationMessage: "This attribute will be removed in a future version of the provider.",
	}
}

func deprecatedWithAlternateMessage(altPath path.Path) string {
	return fmt.Sprintf("Use '%s' instead. This attribute will be removed in a future version of the provider.", altPath.String())
}
