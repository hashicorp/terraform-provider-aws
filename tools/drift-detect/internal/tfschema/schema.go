// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package tfschema extracts and normalises the Terraform provider schema
// produced by `terraform providers schema -json` into an Intermediate
// Representation (IR) that the drift-detect comparison engine can consume.
//
// # Phase 1 scope
//
// LoadFile reads a pre-generated schema JSON file and builds a ProviderSchema
// whose Resources map contains one ResourceIR per resource.  Each ResourceIR
// holds only the root-block, top-level primitive fields (string / int64 /
// float64 / bool) with their Required / Optional / Computed flags.
//
// # Phase 2 – nested structures (NOT YET ACTIVE)
//
// The helpers below that are guarded by the "Phase 2" build tag comments
// implement recursive block and nested-attribute walking.  They are compiled
// but never called from Phase 1 code; uncomment the call sites in
// flattenRootBlock once Phase 2 work begins.
package tfschema

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/zclconf/go-cty/cty"
)

// ---------------------------------------------------------------------------
// IR types – Phase 1
// ---------------------------------------------------------------------------

// FieldType is the normalised primitive type for a schema field.
type FieldType string

const (
	FieldTypeString  FieldType = "string"
	FieldTypeInt64   FieldType = "int64"
	FieldTypeFloat64 FieldType = "float64"
	FieldTypeBool    FieldType = "bool"
	FieldTypeUnknown FieldType = "unknown" // non-primitive; used as a sentinel
)

// Field is the Phase 1 IR for a single schema attribute.
type Field struct {
	Name     string    `json:"name"`
	Type     FieldType `json:"type"`
	Required bool      `json:"required"`
	Optional bool      `json:"optional"`
	Computed bool      `json:"computed"` // read-only, set by provider
}

// Identifier captures resource identity metadata discovered from AWS models.
// It is informational in Phase 1 and is not included in drift comparison.
type Identifier struct {
	Name   string    `json:"name"`
	Type   FieldType `json:"type"`
	Target string    `json:"target,omitempty"`
}

// ResourceMetadata holds optional extraction metadata that is not currently
// part of the drift comparison signal.
type ResourceMetadata struct {
	Identifiers map[string]*Identifier `json:"identifiers,omitempty"`
}

// ResourceIR is the Phase 1 IR for a single Terraform resource schema.
// Fields holds only the root-block, top-level primitive attributes.
type ResourceIR struct {
	// Name is the resource type name, e.g. "aws_s3_bucket".
	Name string
	// Source is always "terraform" for values produced by this package.
	Source string
	// Fields maps normalised attribute name → Field descriptor.
	// Phase 1: root-block primitive attributes only.
	Fields map[string]*Field

	// Metadata carries optional, non-compared information (e.g. identifiers).
	Metadata *ResourceMetadata `json:"metadata,omitempty"`

	// Phase 2 (not yet populated): NestedBlocks will hold recursive block
	// data keyed by dot-path (e.g. "rule", "rule.action").
	// NestedBlocks map[string]*NestedBlock // TODO(phase2): uncomment
}

// ProviderSchema is the top-level container returned by LoadFile.
// Phase 1 populates Resources only; the other maps are reserved for later
// phases or can be populated by extending LoadFile.
type ProviderSchema struct {
	// Resources maps resource type name → ResourceIR.
	Resources map[string]*ResourceIR

	// Phase 2+ (not yet populated).
	// DataSources map[string]*ResourceIR // TODO(phase2): uncomment
	// Ephemerals  map[string]*ResourceIR // TODO(phase2): uncomment
}

// ---------------------------------------------------------------------------
// Phase 2 IR stubs (compiled, not yet called)
// ---------------------------------------------------------------------------

// NestedBlock is the Phase 2 IR for a single schema block at any nesting
// level.  Fields holds the attributes declared directly inside this block;
// ChildBlocks holds the names of its immediate nested-block children.
//
// NOTE: this type is defined now so that ResourceIR.NestedBlocks can be
// un-commented in Phase 2 without touching the type declarations.
type NestedBlock struct {
	// Path is the dot-separated path from the resource root, e.g. "rule.action".
	Path string
	// Fields holds the primitive attributes at this block level.
	Fields map[string]*Field
	// ChildBlocks lists the names of immediate nested-block children.
	ChildBlocks []string
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// LoadFile reads a `terraform providers schema -json` output file and returns
// the parsed ProviderSchema for the given providerSource.
//
// providerSource is the registry address, e.g.
// "registry.terraform.io/hashicorp/aws".  A short name ("aws") is also
// accepted as a fallback match.
func LoadFile(path, providerSource string) (*ProviderSchema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading schema file: %w", err)
	}

	var raw tfjson.ProviderSchemas
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing schema JSON: %w", err)
	}

	provider := findProvider(&raw, providerSource)
	if provider == nil {
		return nil, fmt.Errorf("provider %q not found in schema", providerSource)
	}

	ps := &ProviderSchema{
		Resources: projectResources(provider.ResourceSchemas),
	}
	return ps, nil
}

// ResourceNames returns a sorted slice of all resource type names present in
// the ProviderSchema.  Stable ordering is important for deterministic reports.
func (ps *ProviderSchema) ResourceNames() []string {
	names := make([]string, 0, len(ps.Resources))
	for k := range ps.Resources {
		names = append(names, k)
	}
	slices.Sort(names)
	return names
}

// ---------------------------------------------------------------------------
// Internal projection helpers – Phase 1
// ---------------------------------------------------------------------------

// projectResources converts a raw tfjson resource map into the Phase 1 IR map.
func projectResources(in map[string]*tfjson.Schema) map[string]*ResourceIR {
	out := make(map[string]*ResourceIR, len(in))
	for name, s := range in {
		out[name] = projectResourceIR(name, s)
	}
	return out
}

// projectResourceIR builds a ResourceIR from a single tfjson.Schema.
// Phase 1: only root-block top-level primitive attributes are extracted.
func projectResourceIR(name string, s *tfjson.Schema) *ResourceIR {
	ir := &ResourceIR{
		Name:   name,
		Source: "terraform",
		Fields: make(map[string]*Field),
	}

	if s == nil || s.Block == nil {
		return ir
	}

	flattenRootBlock(ir, s.Block)
	return ir
}

// flattenRootBlock iterates over the top-level attributes of a SchemaBlock and
// adds primitive fields to ir.Fields.
//
// Phase 2 extension points are commented in-line.
func flattenRootBlock(ir *ResourceIR, block *tfjson.SchemaBlock) {
	for attrName, attr := range block.Attributes {
		ft := ctyTypeToFieldType(attr.AttributeType)

		// Phase 1: skip non-primitive attributes (collections, objects, etc.).
		// Phase 2: remove this guard and use FieldTypeUnknown as a marker, or
		// expand with full nested-attribute walking via extractNestedFields().
		if ft == FieldTypeUnknown {
			continue
		}

		ir.Fields[attrName] = &Field{
			Name:     attrName,
			Type:     ft,
			Required: attr.Required,
			Optional: attr.Optional,
			Computed: attr.Computed,
		}
	}

	// Phase 2: recurse into nested blocks.
	// flattenNestedBlocks(ir, "", block)  // TODO(phase2): uncomment
}

// ---------------------------------------------------------------------------
// Phase 2 helpers (compiled, not yet called from Phase 1)
// ---------------------------------------------------------------------------

// flattenNestedBlocks recursively walks tfjson NestedBlocks and populates
// ir.NestedBlocks with a NestedBlock per dot-path.
//
// Call flattenRootBlock → flattenNestedBlocks once Phase 2 begins.
//
//nolint:unused // Phase 2: remove this comment when wired in
func flattenNestedBlocks(ir *ResourceIR, parentPath string, block *tfjson.SchemaBlock) {
	for childName, nb := range block.NestedBlocks {
		childPath := childName
		if parentPath != "" {
			childPath = parentPath + "." + childName
		}

		nested := &NestedBlock{
			Path:   childPath,
			Fields: make(map[string]*Field),
		}

		if nb.Block != nil {
			for attrName, attr := range nb.Block.Attributes {
				ft := ctyTypeToFieldType(attr.AttributeType)
				nested.Fields[attrName] = &Field{
					Name:     attrName,
					Type:     ft,
					Required: attr.Required,
					Optional: attr.Optional,
					Computed: attr.Computed,
				}
			}
			for grandchildName := range nb.Block.NestedBlocks {
				nested.ChildBlocks = append(nested.ChildBlocks, grandchildName)
			}
			// Recurse deeper.
			flattenNestedBlocks(ir, childPath, nb.Block)
		}

		// TODO(phase2): store nested block on ir.NestedBlocks once that field
		// is uncommented on ResourceIR.
		_ = nested
	}
}

// extractNestedFields resolves child attributes from a schema attribute's type
// definition for both framework-style nested attributes (AttributeNestedType)
// and SDK-style cty object types (list/set/map of objects).
//
// Called during Phase 2 attribute-level recursion.
//
//nolint:unused // Phase 2: remove this comment when wired in
func extractNestedFields(attr *tfjson.SchemaAttribute) []*Field {
	// Framework-style: AttributeNestedType with an explicit attributes map.
	if attr.AttributeNestedType != nil && len(attr.AttributeNestedType.Attributes) > 0 {
		return fieldsFromSchemaAttrs(attr.AttributeNestedType.Attributes)
	}

	// SDK-style: cty type encoding (list(object({...})), set(object({...})),
	// object({...})).
	ty := attr.AttributeType
	if ty.IsCollectionType() {
		ty = ty.ElementType()
	}
	if ty.IsObjectType() {
		return fieldsFromCtyObject(ty)
	}

	return nil
}

// fieldsFromSchemaAttrs converts a map of tfjson.SchemaAttribute to []*Field.
//
//nolint:unused // Phase 2
func fieldsFromSchemaAttrs(attrs map[string]*tfjson.SchemaAttribute) []*Field {
	result := make([]*Field, 0, len(attrs))
	for name, a := range attrs {
		result = append(result, &Field{
			Name:     name,
			Type:     ctyTypeToFieldType(a.AttributeType),
			Required: a.Required,
			Optional: a.Optional,
			Computed: a.Computed,
		})
	}
	return result
}

// fieldsFromCtyObject extracts top-level attribute names from a cty object
// type, mapping each to a Field with FieldTypeUnknown (cty objects do not
// carry Required/Optional/Computed; that comes from the enclosing attribute).
//
//nolint:unused // Phase 2
func fieldsFromCtyObject(ty cty.Type) []*Field {
	result := make([]*Field, 0, len(ty.AttributeTypes()))
	for name := range ty.AttributeTypes() {
		result = append(result, &Field{
			Name:     name,
			Type:     FieldTypeUnknown,
			Computed: true, // conservative default for SDK-style objects
		})
	}
	return result
}

// ---------------------------------------------------------------------------
// Type-mapping helpers
// ---------------------------------------------------------------------------

// ctyTypeToFieldType maps a cty.Type to the Phase 1 FieldType set.
// Non-primitive types (collections, objects, dynamic) return FieldTypeUnknown.
func ctyTypeToFieldType(ty cty.Type) FieldType {
	switch ty {
	case cty.String:
		return FieldTypeString
	case cty.Number:
		// Terraform's number type is arbitrary precision; we normalise to
		// int64 for the IR.  Float detection (if needed) can be added in Phase 2.
		return FieldTypeInt64
	case cty.Bool:
		return FieldTypeBool
	default:
		return FieldTypeUnknown
	}
}

// ---------------------------------------------------------------------------
// Utility
// ---------------------------------------------------------------------------

// findProvider looks up a *tfjson.ProviderSchema by its full source address,
// falling back to just the last path segment (e.g. "aws").
func findProvider(ps *tfjson.ProviderSchemas, source string) *tfjson.ProviderSchema {
	if ps.Schemas == nil {
		return nil
	}
	if p, ok := ps.Schemas[source]; ok {
		return p
	}
	// Short-name fallback.
	parts := strings.Split(source, "/")
	short := parts[len(parts)-1]
	if p, ok := ps.Schemas[short]; ok {
		return p
	}
	return nil
}
