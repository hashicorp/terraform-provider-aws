// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package listresourceattribute

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var Region = sync.OnceValue(func() schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Description: names.ListResourceTopLevelRegionAttributeDescription,
	}
})
