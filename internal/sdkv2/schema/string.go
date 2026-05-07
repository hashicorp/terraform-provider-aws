// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"reflect"
	"sync"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/sync"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

type AttrHandling int

const (
	AttrElem     AttrHandling = 0
	AttrRequired AttrHandling = 1 << iota
	AttrOptional
	AttrComputed
	AttrOptionalComputed = AttrOptional | AttrComputed
)

func (x AttrHandling) IsRequired() bool {
	return x&AttrRequired != 0
}

func (x AttrHandling) IsOptional() bool {
	return x&AttrOptional != 0
}

func (x AttrHandling) IsComputed() bool {
	return x&AttrComputed != 0
}

var arnStringSchemaCache tfsync.Map[AttrHandling, *schema.Schema]

func ArnStringSchema(handling AttrHandling) *schema.Schema {
	if handling == AttrComputed {
		return StringComputedOnly()
	}

	s, ok := arnStringSchemaCache.Load(handling)
	if ok {
		return s
	}

	// Use a separate `LoadOrStore` to avoid allocation if item is already in the cache
	// Use `LoadOrStore` instead of `Store` in case there is a race
	s, _ = arnStringSchemaCache.LoadOrStore(
		handling,
		&schema.Schema{
			Type:         schema.TypeString,
			Required:     handling.IsRequired(),
			Optional:     handling.IsOptional(),
			Computed:     handling.IsComputed(),
			ValidateFunc: verify.ValidARN,
		},
	)
	return s
}

func ArnStringDataSourceSchema() *schema.Schema {
	return ArnStringSchema(AttrComputed)
}

type stringEnumIdentity struct {
	handling AttrHandling
	typ      reflect.Type
}

var stringEnumSchemaCache tfsync.Map[stringEnumIdentity, *schema.Schema]

func StringEnumSchema[T enum.Valueser[T]](handling AttrHandling) *schema.Schema {
	if handling == AttrComputed {
		return StringComputedOnly()
	}

	id := stringEnumIdentity{
		handling: handling,
		typ:      reflect.TypeFor[T](),
	}

	s, ok := stringEnumSchemaCache.Load(id)
	if ok {
		return s
	}

	// Use a separate `LoadOrStore` to avoid allocation if item is already in the cache
	// Use `LoadOrStore` instead of `Store` in case there is a race
	s, _ = stringEnumSchemaCache.LoadOrStore(
		id,
		&schema.Schema{
			Type:             schema.TypeString,
			Required:         handling.IsRequired(),
			Optional:         handling.IsOptional(),
			Computed:         handling.IsComputed(),
			ValidateDiagFunc: enum.Validate[T](),
		},
	)
	return s
}

func StringEnumDataSourceSchema[T enum.Valueser[T]]() *schema.Schema {
	return StringEnumSchema[T](AttrComputed)
}

type stringLenBetweenIdentity struct {
	handling AttrHandling
	min, max int
}

var stringLenBetweenSchemaCache tfsync.Map[stringLenBetweenIdentity, *schema.Schema]

func StringLenBetweenSchema(handling AttrHandling, min, max int) *schema.Schema {
	if handling == AttrComputed {
		return StringComputedOnly()
	}

	id := stringLenBetweenIdentity{
		handling: handling,
		min:      min,
		max:      max,
	}

	s, ok := stringLenBetweenSchemaCache.Load(id)
	if ok {
		return s
	}

	// Use a separate `LoadOrStore` to avoid allocation if item is already in the cache
	// Use `LoadOrStore` instead of `Store` in case there is a race
	s, _ = stringLenBetweenSchemaCache.LoadOrStore(
		id,
		&schema.Schema{
			Type:         schema.TypeString,
			Required:     handling.IsRequired(),
			Optional:     handling.IsOptional(),
			Computed:     handling.IsComputed(),
			ValidateFunc: validation.StringLenBetween(min, max),
		},
	)
	return s
}

type stringMatchIdentity struct {
	handling    AttrHandling
	re, message string
}

var stringMatchSchemaCache tfsync.Map[stringMatchIdentity, *schema.Schema]

func StringMatchSchema(handling AttrHandling, re, message string) *schema.Schema {
	if handling == AttrComputed {
		return StringComputedOnly()
	}

	id := stringMatchIdentity{
		handling: handling,
		re:       re,
		message:  message,
	}

	s, ok := stringMatchSchemaCache.Load(id)
	if ok {
		return s
	}

	// Use a separate `LoadOrStore` to avoid allocation if item is already in the cache
	// Use `LoadOrStore` instead of `Store` in case there is a race
	s, _ = stringMatchSchemaCache.LoadOrStore(
		id,
		&schema.Schema{
			Type:         schema.TypeString,
			Required:     handling.IsRequired(),
			Optional:     handling.IsOptional(),
			Computed:     handling.IsComputed(),
			ValidateFunc: validation.StringMatch(regexache.MustCompile(re), message),
		},
	)
	return s
}

var utcTimestampStringSchemaCache tfsync.Map[AttrHandling, *schema.Schema]

func UTCTimestampStringSchema(handling AttrHandling) *schema.Schema {
	s, ok := utcTimestampStringSchemaCache.Load(handling)
	if ok {
		return s
	}

	// Use a separate `LoadOrStore` to avoid allocation if item is already in the cache
	// Use `LoadOrStore` instead of `Store` in case there is a race
	s, _ = utcTimestampStringSchemaCache.LoadOrStore(
		handling,
		&schema.Schema{
			Type:         schema.TypeString,
			Required:     handling.IsRequired(),
			Optional:     handling.IsOptional(),
			Computed:     handling.IsComputed(),
			ValidateFunc: verify.ValidUTCTimestamp,
		},
	)
	return s
}

var StringComputedOnly = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	}
})
