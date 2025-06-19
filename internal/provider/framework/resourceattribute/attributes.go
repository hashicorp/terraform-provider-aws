// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourceattribute

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var Region = sync.OnceValue(func() schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: names.TopLevelRegionAttributeDescription,
	}
})
