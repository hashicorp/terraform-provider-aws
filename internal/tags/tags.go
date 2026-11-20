// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package tags

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
)

// TagsSchema returns the schema to use for configurable resource tags.
var TagsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	}
})

// TagsSchemaOptionalComputedWarnShouldBeComputedOnly returns the schema for existing tags that should be
// Computed only but were made Optional+Computed by mistake: resource `tags_all`, and data source `tags`
// that Read doesn't use. The old helper's name, TagsSchemaComputed, suggested Computed only.
// Validation warns when the value is configured. It stays Optional so existing configurations keep
// working until a future version makes these Computed only
// (https://github.com/hashicorp/terraform-provider-aws/issues/50581).
//
// Don't use it in new code. Use TagsSchemaComputedOnly, or TagsSchemaOptionalComputed for a data
// source that filters on tags.
var TagsSchemaOptionalComputedWarnShouldBeComputedOnly = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeMap,
		Optional:         true,
		Computed:         true,
		Elem:             &schema.Schema{Type: schema.TypeString},
		ValidateDiagFunc: sdkv2.WarnShouldBeComputedOnly,
	}
})

// TagsSchemaComputedOnly returns the schema for tags that only the provider sets: new resource
// `tags_all`, new data source `tags` that Read doesn't use, and tags nested in blocks users can't
// configure. Users can't set it in configuration.
var TagsSchemaComputedOnly = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Computed: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	}
})

// TagsSchemaOptionalComputed returns the schema to use for data source tags that are both
// a filter argument and an attribute. Use it only where Read uses the configured value,
// for example to build EC2 tag filters.
var TagsSchemaOptionalComputed = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Computed: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	}
})

// TagsSchemaForceNew returns the schema to use for configurable resource tags where changes recreate the resource.
var TagsSchemaForceNew = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		ForceNew: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	}
})
