// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tags

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TagsSchema returns the schema to use for configurable resource tags.
var TagsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	}
})

// TagsSchema returns the schema to use for computed tags, either resource `tags_all` or for data sources.
var TagsSchemaComputed = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Computed: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	}
})

// TagsSchema returns the schema to use for configurable resource tags where changes recreate the resource.
var TagsSchemaForceNew = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		ForceNew: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	}
})
