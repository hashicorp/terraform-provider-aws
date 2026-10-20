// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package tags

import (
	"sync"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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

// TagsSchemaComputed returns the schema to use for resource `tags_all` and other resource tags
// that the provider computes. Validation warns when it's configured; it stays Optional for now so
// that existing configurations keep working.
var TagsSchemaComputed = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeMap,
		Optional:         true,
		Computed:         true,
		Elem:             &schema.Schema{Type: schema.TypeString},
		ValidateDiagFunc: warnComputedTagsConfigured,
	}
})

func warnComputedTagsConfigured(any, cty.Path) diag.Diagnostics {
	return diag.Diagnostics{{
		Severity: diag.Warning,
		Summary:  "Argument is computed",
		Detail:   "The provider sets this attribute. Remove it from your configuration. A future version of the provider will reject it.",
	}}
}

// TagsSchemaComputedOnly returns the schema to use for data source tags that are attributes only.
// Users can't set it in configuration.
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
