// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/sync"
)

type floatBetweenIdentity struct {
	handling AttrHandling
	min, max float64
}

var floatBetweenSchemaCache tfsync.Map[floatBetweenIdentity, *schema.Schema]

func FloatBetweenSchema(handling AttrHandling, min, max float64) *schema.Schema {
	id := floatBetweenIdentity{
		handling: handling,
		min:      min,
		max:      max,
	}

	s, ok := floatBetweenSchemaCache.Load(id)
	if ok {
		return s
	}

	// Use a separate `LoadOrStore` to avoid allocation if item is already in the cache
	// Use `LoadOrStore` instead of `Store` in case there is a race
	s, _ = floatBetweenSchemaCache.LoadOrStore(
		id,
		&schema.Schema{
			Type:         schema.TypeFloat,
			Required:     handling.IsRequired(),
			Optional:     handling.IsOptional(),
			Computed:     handling.IsComputed(),
			ValidateFunc: validation.FloatBetween(min, max),
		},
	)
	return s
}

var FloatComputedOnly = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeFloat,
		Computed: true,
	}
})
