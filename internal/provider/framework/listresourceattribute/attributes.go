// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package listresourceattribute

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var Region = sync.OnceValue(func() schema.StringAttribute {
	return schema.StringAttribute{
		Optional:    true,
		Description: names.ListResourceTopLevelRegionAttributeDescription,
	}
})
