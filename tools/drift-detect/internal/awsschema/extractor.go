// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package awsschema extracts an AWS Intermediate Representation (IR) from
// Smithy 2.0 model files using the resource mapping configuration.
//
// The extracted IR uses the same ResourceIR / Field types as internal/tfschema
// so the comparison engine can diff the two sides without any translation.
//
// Extraction is capability-based:
//   - lifecycle extraction from Smithy operations (inferred from smithy_resource
//     and/or overridden by lifecycle mappings),
//   - enum expansion from attribute_map_enum,
//   - explicit field additions from explicit_fields.
package awsschema

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/tools/drift-detect/internal/awsmapping"
	"github.com/hashicorp/terraform-provider-aws/tools/drift-detect/internal/smithymodel"
	"github.com/hashicorp/terraform-provider-aws/tools/drift-detect/internal/tfschema"
)

// ExtractResource builds a ResourceIR for the given Terraform resource name
// using the provided mapping and the api-models raw base URL.
//
// apiModelsBaseURL is an HTTP(S) base URL such as
// https://raw.githubusercontent.com/aws/api-models-aws/main. The
// smithy_model path in the mapping is joined to this base URL.
func ExtractResource(tfName string, m *awsmapping.ResourceMapping, apiModelsBaseURL string) (*tfschema.ResourceIR, error) {
	ir := &tfschema.ResourceIR{
		Name:   tfName,
		Source: "aws",
		Fields: make(map[string]*tfschema.Field),
	}

	if !hasAnyExtractionConfig(m) {
		return nil, fmt.Errorf(
			"resource %q has no extraction configuration; set one or more of smithy_resource/lifecycle, attribute_map_enum, explicit_fields",
			tfName,
		)
	}

	var model *smithymodel.Model
	if needsSmithyModel(m) {
		if m.SmithyModel == "" {
			return nil, fmt.Errorf("resource %q requires smithy_model for lifecycle/enum extraction", tfName)
		}

		var err error
		model, err = loadModel(apiModelsBaseURL, m.SmithyModel)
		if err != nil {
			return nil, fmt.Errorf("loading smithy model: %w", err)
		}
	}

	if model != nil {
		if hasLifecycleExtraction(m) {
			lifecycle, err := resolveLifecycle(m, model)
			if err != nil {
				return nil, err
			}
			extractLifecycle(ir, m, model, lifecycle)
		}

		setIdentifierMetadata(ir, m, model)

		if hasEnumExtraction(m) {
			addEnumFields(ir, m, model)
		}
	}

	addExplicitFields(ir, m)

	return ir, nil
}

func hasLifecycleExtraction(m *awsmapping.ResourceMapping) bool {
	if m.SmithyResource != "" {
		return true
	}

	lc := m.Lifecycle
	return lc.Create != "" || lc.Put != "" || lc.Read != "" || lc.Update != "" || lc.Delete != "" || lc.List != ""
}

func hasEnumExtraction(m *awsmapping.ResourceMapping) bool {
	return m.AttributeMapEnum != ""
}

func hasExplicitExtraction(m *awsmapping.ResourceMapping) bool {
	return len(m.ExplicitFields) > 0
}

func needsSmithyModel(m *awsmapping.ResourceMapping) bool {
	return hasLifecycleExtraction(m) || hasEnumExtraction(m) || m.SmithyResource != ""
}

func hasAnyExtractionConfig(m *awsmapping.ResourceMapping) bool {
	return hasLifecycleExtraction(m) || hasEnumExtraction(m) || hasExplicitExtraction(m)
}

func loadModel(apiModelsBaseURL, smithyModelPath string) (*smithymodel.Model, error) {
	modelURL, err := url.JoinPath(strings.TrimRight(apiModelsBaseURL, "/"), smithyModelPath)
	if err != nil {
		return nil, fmt.Errorf("joining smithy model URL: %w", err)
	}

	return smithymodel.LoadURL(modelURL)
}

func resolveLifecycle(m *awsmapping.ResourceMapping, model *smithymodel.Model) (awsmapping.Lifecycle, error) {
	resolved := awsmapping.Lifecycle{}

	if m.SmithyResource != "" {
		resourceID := m.SmithyNamespace + "#" + m.SmithyResource
		shape := model.Shape(resourceID)
		if shape == nil || shape.Kind != smithymodel.KindResource {
			return resolved, fmt.Errorf("resource shape %q not found", resourceID)
		}
		resolved = awsmapping.Lifecycle{
			Create: opName(shape.CreateTarget),
			Put:    opName(shape.PutTarget),
			Read:   opName(shape.ReadTarget),
			Update: opName(shape.UpdateTarget),
			Delete: opName(shape.DeleteTarget),
			List:   opName(shape.ListTarget),
		}
	}

	// Explicit mapping values override inferred values.
	if m.Lifecycle.Create != "" {
		resolved.Create = m.Lifecycle.Create
	}
	if m.Lifecycle.Put != "" {
		resolved.Put = m.Lifecycle.Put
	}
	if m.Lifecycle.Read != "" {
		resolved.Read = m.Lifecycle.Read
	}
	if m.Lifecycle.Update != "" {
		resolved.Update = m.Lifecycle.Update
	}
	if m.Lifecycle.Delete != "" {
		resolved.Delete = m.Lifecycle.Delete
	}
	if m.Lifecycle.List != "" {
		resolved.List = m.Lifecycle.List
	}

	return resolved, nil
}

func setIdentifierMetadata(ir *tfschema.ResourceIR, m *awsmapping.ResourceMapping, model *smithymodel.Model) {
	if m.SmithyResource == "" {
		return
	}

	resourceID := m.SmithyNamespace + "#" + m.SmithyResource
	shape := model.Shape(resourceID)
	if shape == nil || shape.Kind != smithymodel.KindResource || len(shape.Identifiers) == 0 {
		return
	}

	identifiers := make(map[string]*tfschema.Identifier, len(shape.Identifiers))
	for awsName, target := range shape.Identifiers {
		tfName := m.TFName(awsName)
		identifiers[tfName] = &tfschema.Identifier{
			Name:   tfName,
			Type:   resolveFieldType(model, target),
			Target: target,
		}
	}

	ir.Metadata = &tfschema.ResourceMetadata{Identifiers: identifiers}
}

func opName(target string) string {
	if target == "" {
		return ""
	}
	_, name, ok := strings.Cut(target, "#")
	if !ok {
		return target
	}
	return name
}

// ---------------------------------------------------------------------------
// Lifecycle extraction
// ---------------------------------------------------------------------------

func extractLifecycle(ir *tfschema.ResourceIR, m *awsmapping.ResourceMapping, model *smithymodel.Model, lifecycle awsmapping.Lifecycle) {
	// Union fields from write operation inputs and read/list outputs.
	ops := []string{
		lifecycle.Create,
		lifecycle.Put,
		lifecycle.Read,
		lifecycle.Update,
		lifecycle.Delete,
		lifecycle.List,
	}

	seenOps := map[string]bool{}
	for _, opName := range ops {
		if opName == "" || seenOps[opName] {
			continue
		}
		seenOps[opName] = true

		opID := m.SmithyNamespace + "#" + opName
		op := model.Shape(opID)
		if op == nil || op.Kind != smithymodel.KindOperation {
			continue
		}

		if op.InputTarget != "" {
			addStructureFields(ir, m, model, op.InputTarget)
		}
		if op.OutputTarget != "" {
			addStructureFields(ir, m, model, op.OutputTarget)
		}
	}
}

// addStructureFields walks a structure shape and adds its primitive members
// to the IR, skipping suppressed and transport fields.
func addStructureFields(ir *tfschema.ResourceIR, m *awsmapping.ResourceMapping, model *smithymodel.Model, structID string) {
	s := model.Shape(structID)
	if s == nil || s.Kind != smithymodel.KindStructure {
		return
	}

	for memberName, mem := range s.Members {

		// Skip transport/identity fields detected by traits.
		if mem.Traits.IsSuppressible() {
			continue
		}
		// Skip fields listed in suppress_fields of the mapping.
		if m.IsSuppressed(memberName) {
			continue
		}

		ft := resolveFieldType(model, mem.Target)
		if ft == tfschema.FieldTypeUnknown {
			// Phase 1: skip non-primitive types.
			continue
		}

		tfName := m.TFName(memberName)
		// Don't overwrite a field already set from a previous operation
		// (create wins over read for Required semantics).
		if _, exists := ir.Fields[tfName]; !exists {
			ir.Fields[tfName] = &tfschema.Field{
				Name:     tfName,
				Type:     ft,
				Required: mem.Traits.Required,
				Optional: !mem.Traits.Required,
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Enum expansion
// ---------------------------------------------------------------------------

func addEnumFields(ir *tfschema.ResourceIR, m *awsmapping.ResourceMapping, model *smithymodel.Model) {
	enumID := m.SmithyNamespace + "#" + m.AttributeMapEnum
	enumShape := model.Shape(enumID)
	if enumShape == nil || enumShape.Kind != smithymodel.KindEnum {
		return
	}

	for _, mem := range enumShape.Members {
		awsName := mem.EnumValue
		if awsName == "" {
			continue
		}
		if m.IsSuppressed(awsName) {
			continue
		}
		tfName := m.TFName(awsName)
		if _, exists := ir.Fields[tfName]; !exists {
			// Enum-backed attribute maps carry string values.
			// Required is unknown at the model level — default to optional.
			ir.Fields[tfName] = &tfschema.Field{
				Name:     tfName,
				Type:     tfschema.FieldTypeString,
				Required: false,
				Optional: true,
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Explicit fields from mapping
// ---------------------------------------------------------------------------

func addExplicitFields(ir *tfschema.ResourceIR, m *awsmapping.ResourceMapping) {
	for _, ef := range m.ExplicitFields {
		tfName := m.TFName(ef.Name)
		if _, exists := ir.Fields[tfName]; exists {
			continue
		}
		ft := explicitFieldType(ef.Type)
		ir.Fields[tfName] = &tfschema.Field{
			Name:     tfName,
			Type:     ft,
			Required: ef.Required,
			Optional: !ef.Required,
		}
	}
}

// ---------------------------------------------------------------------------
// Type resolution
// ---------------------------------------------------------------------------

// resolveFieldType resolves a Smithy shape target ID to a Phase 1 FieldType.
// Non-primitive shapes (structures, lists, maps, unions) return FieldTypeUnknown.
func resolveFieldType(model *smithymodel.Model, targetID string) tfschema.FieldType {
	kind := model.ResolveToKind(targetID)
	return smithyKindToFieldType(kind)
}

// smithyKindToFieldType maps a Smithy ShapeKind to the Phase 1 IR FieldType.
func smithyKindToFieldType(k smithymodel.ShapeKind) tfschema.FieldType {
	switch k {
	case smithymodel.KindString, smithymodel.KindBlob,
		smithymodel.KindTimestamp, smithymodel.KindDocument:
		return tfschema.FieldTypeString
	case smithymodel.KindBoolean:
		return tfschema.FieldTypeBool
	case smithymodel.KindInteger, smithymodel.KindLong,
		smithymodel.KindByte, smithymodel.KindShort,
		smithymodel.KindBigInteger:
		return tfschema.FieldTypeInt64
	case smithymodel.KindFloat, smithymodel.KindDouble,
		smithymodel.KindBigDecimal:
		return tfschema.FieldTypeFloat64
	default:
		return tfschema.FieldTypeUnknown
	}
}

// explicitFieldType converts the string type name in an ExplicitField entry
// to a FieldType.
func explicitFieldType(s string) tfschema.FieldType {
	switch s {
	case "string":
		return tfschema.FieldTypeString
	case "bool", "boolean":
		return tfschema.FieldTypeBool
	case "int64", "int", "integer":
		return tfschema.FieldTypeInt64
	case "float64", "float":
		return tfschema.FieldTypeFloat64
	default:
		return tfschema.FieldTypeString // safe default for explicit fields
	}
}
