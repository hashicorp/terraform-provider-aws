// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/sync"
)

type intAtLeastIdentity struct {
	handling AttrHandling
	min      int
}

var intAtLeastSchemaCache tfsync.Map[intAtLeastIdentity, *schema.Schema]

func IntAtLeastSchema(handling AttrHandling, min int) *schema.Schema {
	id := intAtLeastIdentity{
		handling: handling,
		min:      min,
	}

	s, ok := intAtLeastSchemaCache.Load(id)
	if ok {
		return s
	}

	// Use a separate `LoadOrStore` to avoid allocation if item is already in the cache
	// Use `LoadOrStore` instead of `Store` in case there is a race
	s, _ = intAtLeastSchemaCache.LoadOrStore(
		id,
		&schema.Schema{
			Type:         schema.TypeInt,
			Required:     handling.IsRequired(),
			Optional:     handling.IsOptional(),
			Computed:     handling.IsComputed(),
			ValidateFunc: validation.IntAtLeast(min),
		},
	)
	return s
}

type intBetweenIdentity struct {
	handling AttrHandling
	min, max int
}

var intBetweenSchemaCache tfsync.Map[intBetweenIdentity, *schema.Schema]

func IntBetweenSchema(handling AttrHandling, min, max int) *schema.Schema {
	id := intBetweenIdentity{
		handling: handling,
		min:      min,
		max:      max,
	}

	s, ok := intBetweenSchemaCache.Load(id)
	if ok {
		return s
	}

	// Use a separate `LoadOrStore` to avoid allocation if item is already in the cache
	// Use `LoadOrStore` instead of `Store` in case there is a race
	s, _ = intBetweenSchemaCache.LoadOrStore(
		id,
		&schema.Schema{
			Type:         schema.TypeInt,
			Required:     handling.IsRequired(),
			Optional:     handling.IsOptional(),
			Computed:     handling.IsComputed(),
			ValidateFunc: validation.IntBetween(min, max),
		},
	)
	return s
}

var IntComputedOnly = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Computed: true,
	}
})
