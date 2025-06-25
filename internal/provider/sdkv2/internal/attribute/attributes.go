// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package attribute

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var Region = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: names.TopLevelRegionAttributeDescription,
	}
})
