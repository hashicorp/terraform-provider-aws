// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// ComputedOnlyFromSchema is a recursive function that converts an
// existing Resource schema to a computed-only schema
//
// This may be used to create computed-only variants of a resource attribute,
// or to copy a resource schema into a corresponding data source. All schema
// elements are copied, but certain attributes are ignored or changed:
// - All attributes have Computed = true
// - All attributes have ForceNew, Required = false
// - Validation functions and attributes (e.g. MaxItems) are not copied
//
// Adapted from https://github.com/hashicorp/terraform-provider-google/google/datasource_helpers.go. Thanks!
func ComputedOnlyFromSchema(s *schema.Schema) *schema.Schema {
	cs := &schema.Schema{
		Computed:    true,
		Description: s.Description,
		Type:        s.Type,
	}

	switch s.Type {
	case schema.TypeSet:
		cs.Set = s.Set
		fallthrough
	case schema.TypeList, schema.TypeMap:
		// List & Set types are generally used for 2 cases:
		// - a list/set of simple primitive values (e.g. list of strings)
		// - a sub resource
		// Maps are usually used for maps of simple primitives
		switch elem := s.Elem.(type) {
		case *schema.Resource:
			// handle the case where the Element is a sub-resource
			cs.Elem = ComputedOnlyFromResource(elem)
		case *schema.Schema:
			// handle simple primitive case
			cs.Elem = &schema.Schema{Type: elem.Type}
		}
	}

	return cs
}

func ComputedOnlyFromResource(r *schema.Resource) *schema.Resource {
	cs := &schema.Resource{
		Schema: ComputedOnlyFromResourceSchema(r.Schema),
	}

	return cs
}

func ComputedOnlyFromResourceSchema(rs map[string]*schema.Schema) map[string]*schema.Schema {
	cs := make(map[string]*schema.Schema, len(rs))

	for k, v := range rs {
		cs[k] = ComputedOnlyFromSchema(v)
	}

	return cs
}

// JSONDocumentSchemaRequired returns the standard schema for a required JSON document.
var JSONDocumentSchemaRequired = sync.OnceValue(jsonDocumentSchemaRequiredFunc(SuppressEquivalentJSONDocuments))

// IAMPolicyDocumentSchemaRequired returns the standard schema for a required IAM policy JSON document.
var IAMPolicyDocumentSchemaRequired = sync.OnceValue(jsonDocumentSchemaRequiredFunc(SuppressEquivalentIAMPolicyDocuments))

func jsonDocumentSchemaRequiredFunc(diffSuppressFunc schema.SchemaDiffSuppressFunc) func() *schema.Schema {
	return func() *schema.Schema {
		return &schema.Schema{
			Type:                  schema.TypeString,
			Required:              true,
			ValidateFunc:          validation.StringIsJSON,
			DiffSuppressFunc:      diffSuppressFunc,
			DiffSuppressOnRefresh: true,
			StateFunc:             NormalizeJsonStringSchemaStateFunc,
		}
	}
}
