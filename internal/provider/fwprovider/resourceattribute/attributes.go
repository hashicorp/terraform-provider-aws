// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourceattribute

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var Region = sync.OnceValue(func() schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: `The AWS Region to use for API operations. Overrides the Region set in the provider configuration.`,
	}
})
