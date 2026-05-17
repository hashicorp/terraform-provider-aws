// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resourceattribute

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var Region = sync.OnceValue(func() schema.StringAttribute {
	return schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: names.ResourceTopLevelRegionAttributeDescription,
	}
})

var RegionDeprecated = sync.OnceValue(func() schema.StringAttribute {
	return schema.StringAttribute{
		Optional:           true,
		Computed:           true,
		Description:        names.ResourceTopLevelRegionAttributeDescription,
		DeprecationMessage: "This attribute will be removed in a future version of the provider.",
	}
})
