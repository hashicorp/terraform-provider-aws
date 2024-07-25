// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	"github.com/hashicorp/terraform-provider-aws/tools/tfsdk2fw/naming"
	"golang.org/x/exp/slices"
)

var (
	dataSourceType = flag.String("data-source", "", "Data Source type")
	resourceType   = flag.String("resource", "", "Resource type")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\ttfsdk2fw [-resource <resource-type>|-data-source <data-source-type>] <package-name> <name> <generated-file>\n\n")
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if len(args) < 3 || (*dataSourceType == "" && *resourceType == "") {
		flag.Usage()
		os.Exit(2)
	}

	packageName := args[0]
	name := args[1]
	outputFilename := args[2]

	// ui := &cli.BasicUi{
	// 	Reader:      os.Stdin,
	// 	Writer:      os.Stdout,
	// 	ErrorWriter: os.Stderr,
	// }
	g := common.NewGenerator()
	migrator := &migrator{
		Generator:   g,
		Name:        name,
		PackageName: packageName,
	}

	p, err := provider.New(context.Background())

	if err != nil {
		g.Fatalf(err.Error())
	}

	if v := *dataSourceType; v != "" {
		resource, ok := p.DataSourcesMap[v]

		if !ok {
			g.Fatalf("data source type %s not found", v)
		}

		migrator.IsDataSource = true
		migrator.Resource = resource
		migrator.Template = datasourceImpl
		migrator.TFTypeName = v
	} else if v := *resourceType; v != "" {
		resource, ok := p.ResourcesMap[v]

		if !ok {
			g.Fatalf("resource type %s not found", v)
		}

		migrator.Resource = resource
		migrator.Template = resourceImpl
		migrator.TFTypeName = v
	}

	if err := migrator.migrate(outputFilename); err != nil {
		g.Fatalf("error migrating Terraform %s schema: %s", *resourceType, err)
	}
}

type migrator struct {
	Generator    *common.Generator
	IsDataSource bool
	Name         string
	PackageName  string
	Resource     *schema.Resource
	Template     string
	TFTypeName   string
}

// migrate generates an identical schema into the specified output file.
func (m *migrator) migrate(outputFilename string) error {
	m.infof("generating into %[1]q", outputFilename)

	// Create target directory.
	dirname := path.Dir(outputFilename)
	err := os.MkdirAll(dirname, 0755)

	if err != nil {
		return fmt.Errorf("creating target directory %s: %w", dirname, err)
	}

	templateData, err := m.generateTemplateData()

	if err != nil {
		return err
	}

	d := m.Generator.NewGoFileDestination(outputFilename)

	if err := d.WriteTemplate("schema", m.Template, templateData); err != nil {
		return err
	}

	return d.Write()
}

func (m *migrator) generateTemplateData() (*templateData, error) {
	sbSchema := strings.Builder{}
	sbStruct := strings.Builder{}
	emitter := &emitter{
		Generator:    m.Generator,
		IsDataSource: m.IsDataSource,
		SchemaWriter: &sbSchema,
		StructWriter: &sbStruct,
	}

	err := emitter.emitSchemaForResource(m.Resource)

	if err != nil {
		return nil, fmt.Errorf("emitting schema code: %w", err)
	}

	templateData := &templateData{
		DefaultCreateTimeout:         emitter.DefaultCreateTimeout,
		DefaultReadTimeout:           emitter.DefaultReadTimeout,
		DefaultUpdateTimeout:         emitter.DefaultUpdateTimeout,
		DefaultDeleteTimeout:         emitter.DefaultDeleteTimeout,
		EmitResourceImportState:      m.Resource.Importer != nil,
		EmitResourceModifyPlan:       !m.IsDataSource && emitter.HasTopLevelTagsAllMap && emitter.HasTopLevelTagsMap,
		EmitResourceUpdateSkeleton:   m.Resource.Update != nil || m.Resource.UpdateContext != nil || m.Resource.UpdateWithoutTimeout != nil,
		HasTimeouts:                  emitter.HasTimeouts,
		ImportFrameworkAttr:          emitter.ImportFrameworkAttr,
		ImportProviderFrameworkTypes: emitter.ImportProviderFrameworkTypes,
		Name:                         m.Name,
		PackageName:                  m.PackageName,
		Schema:                       sbSchema.String(),
		Struct:                       sbStruct.String(),
		TFTypeName:                   m.TFTypeName,
	}

	for _, v := range emitter.FrameworkPlanModifierPackages {
		if !slices.Contains(templateData.FrameworkPlanModifierPackages, v) {
			templateData.FrameworkPlanModifierPackages = append(templateData.FrameworkPlanModifierPackages, v)
		}
	}
	for _, v := range emitter.FrameworkValidatorsPackages {
		if !slices.Contains(templateData.FrameworkValidatorsPackages, v) {
			templateData.FrameworkValidatorsPackages = append(templateData.FrameworkValidatorsPackages, v)
		}
	}
	for _, v := range emitter.GoImports {
		if !slices.Contains(templateData.GoImports, v) {
			templateData.GoImports = append(templateData.GoImports, v)
		}
	}

	return templateData, nil
}

func (m *migrator) infof(format string, a ...interface{}) {
	m.Generator.Infof(format, a...)
}

type emitter struct {
	DefaultCreateTimeout          int64
	DefaultReadTimeout            int64
	DefaultUpdateTimeout          int64
	DefaultDeleteTimeout          int64
	Generator                     *common.Generator
	FrameworkPlanModifierPackages []string // Package names for any terraform-plugin-framework plan modifiers. May contain duplicates.
	FrameworkValidatorsPackages   []string // Package names for any terraform-plugin-framework-validators validators. May contain duplicates.
	GoImports                     []goImport
	HasTimeouts                   bool
	HasTopLevelTagsAllMap         bool
	HasTopLevelTagsMap            bool
	ImportFrameworkAttr           bool
	ImportProviderFrameworkTypes  bool
	IsDataSource                  bool
	SchemaWriter                  io.Writer
	StructWriter                  io.Writer
}

// emitSchemaForResource generates the Plugin Framework code for a Plugin SDK Resource and emits the generated code to the emitter's Writer.
func (e *emitter) emitSchemaForResource(resource *schema.Resource) error {
	if _, ok := resource.Schema["id"]; ok {
		e.warnf("Explicit `id` attribute defined")
	} else {
		resource.Schema["id"] = &schema.Schema{
			Type:     schema.TypeString,
			Optional: e.IsDataSource,
			Computed: true,
		}
	}

	if v := resource.Timeouts; v != nil {
		e.HasTimeouts = true

		if v := v.Create; v != nil {
			e.DefaultCreateTimeout = int64(*v)
		}
		if v := v.Read; v != nil {
			e.DefaultReadTimeout = int64(*v)
		}
		if v := v.Update; v != nil {
			e.DefaultUpdateTimeout = int64(*v)
		}
		if v := v.Delete; v != nil {
			e.DefaultDeleteTimeout = int64(*v)
		}
	}

	fprintf(e.SchemaWriter, "schema.Schema{\n")

	err := e.emitAttributesAndBlocks(nil, resource.Schema)

	if err != nil {
		return err
	}

	if version := resource.SchemaVersion; version > 0 {
		fprintf(e.SchemaWriter, "Version:%d,\n", version)
	}

	if description := resource.Description; description != "" {
		fprintf(e.SchemaWriter, "Description:%q,\n", description)
	}

	if deprecationMessage := resource.DeprecationMessage; deprecationMessage != "" {
		fprintf(e.SchemaWriter, "DeprecationMessage:%q,\n", deprecationMessage)
	}

	fprintf(e.SchemaWriter, "}")

	return nil
}

// emitAttributesAndBlocks generates the Plugin Framework code for a set of Plugin SDK Attributes and Blocks
// and emits the generated code to the emitter's Writer.
// Property names are sorted prior to code generation to reduce diffs.
func (e *emitter) emitAttributesAndBlocks(path []string, schema map[string]*schema.Schema) error {
	isTopLevelAttribute := len(path) == 0

	// At this point we are emitting code for a schema.Block or Schema.
	names := make([]string, 0)
	for name := range schema {
		names = append(names, name)
	}
	sort.Strings(names)

	emittedFieldName := false
	for _, name := range names {
		property := schema[name]

		if !isAttribute(property) {
			continue
		}

		if !emittedFieldName {
			fprintf(e.SchemaWriter, "Attributes: map[string]schema.Attribute{\n")
			emittedFieldName = true
		}

		fprintf(e.SchemaWriter, "%q:", name)

		if isTopLevelAttribute {
			fprintf(e.StructWriter, "%s ", naming.ToCamelCase(name))
		}

		err := e.emitAttributeProperty(append(path, name), property)

		if err != nil {
			return err
		}

		if isTopLevelAttribute {
			fprintf(e.StructWriter, " `tfsdk:%q`\n", name)
		}

		fprintf(e.SchemaWriter, ",\n")
	}
	if emittedFieldName {
		fprintf(e.SchemaWriter, "},\n")
	}

	emittedFieldName = false
	for _, name := range names {
		property := schema[name]

		if isAttribute(property) {
			continue
		}

		if !emittedFieldName {
			fprintf(e.SchemaWriter, "Blocks: map[string]schema.Block{\n")
			emittedFieldName = true
		}

		fprintf(e.SchemaWriter, "%q:", name)

		err := e.emitBlockProperty(append(path, name), property)

		if err != nil {
			return err
		}

		fprintf(e.SchemaWriter, ",\n")
	}
	if emittedFieldName {
		fprintf(e.SchemaWriter, "},\n")
	}

	return nil
}

// emitAttributeProperty generates the Plugin Framework code for a Plugin SDK Attribute's property
// and emits the generated code to the emitter's Writer.
func (e *emitter) emitAttributeProperty(path []string, property *schema.Schema) error {
	attributeName := path[len(path)-1]
	isComputedOnly := property.Computed && !property.Optional
	isTopLevelAttribute := len(path) == 1
	var planModifiers []string
	var defaultSpec string
	var fwPlanModifierPackage, fwPlanModifierType, fwValidatorsPackage, fwValidatorType string

	// At this point we are emitting code for the values of a schema.Schema's Attributes (map[string]schema.Attribute).
	switch v := property.Type; v {
	//
	// Primitive types.
	//
	case schema.TypeBool:
		fprintf(e.SchemaWriter, "schema.BoolAttribute{\n")

		if isTopLevelAttribute {
			fprintf(e.StructWriter, "types.Bool")
		}

		fwPlanModifierPackage = "boolplanmodifier"
		fwPlanModifierType = "Bool"

	case schema.TypeFloat:
		fprintf(e.SchemaWriter, "schema.Float64Attribute{\n")

		if isTopLevelAttribute {
			fprintf(e.StructWriter, "types.Float64")
		}

		fwPlanModifierPackage = "float64planmodifier"
		fwPlanModifierType = "Float64"

	case schema.TypeInt:
		fprintf(e.SchemaWriter, "schema.Int64Attribute{\n")

		if isTopLevelAttribute {
			fprintf(e.StructWriter, "types.Int64")
		}

		fwPlanModifierPackage = "int64planmodifier"
		fwPlanModifierType = "Int64"

	case schema.TypeString:
		// Computed-only ARN attributes are easiest handled as strings.
		if (attributeName == "arn" || strings.HasSuffix(attributeName, "_arn")) && !isComputedOnly {
			e.ImportProviderFrameworkTypes = true

			fprintf(e.SchemaWriter, "schema.StringAttribute{\n")
			fprintf(e.SchemaWriter, "CustomType:fwtypes.ARNType,\n")

			if isTopLevelAttribute {
				fprintf(e.StructWriter, "fwtypes.ARN")
			}
		} else {
			if isTopLevelAttribute && attributeName == "id" {
				fprintf(e.SchemaWriter, "// TODO framework.IDAttribute()\n")
			}

			fprintf(e.SchemaWriter, "schema.StringAttribute{\n")

			if isTopLevelAttribute {
				fprintf(e.StructWriter, "types.String")
			}
		}

		fwPlanModifierPackage = "stringplanmodifier"
		fwPlanModifierType = "String"

	//
	// Complex types.
	//
	case schema.TypeList, schema.TypeMap, schema.TypeSet:
		var aggregateSchemaFactory, typeName string

		switch v {
		case schema.TypeList:
			aggregateSchemaFactory = "schema.ListAttribute{"
			typeName = "list"

			if isTopLevelAttribute {
				fprintf(e.StructWriter, "types.List")
			}

			fwPlanModifierPackage = "listplanmodifier"
			fwPlanModifierType = "List"
			fwValidatorsPackage = "listvalidator"
			fwValidatorType = "List"

		case schema.TypeMap:
			aggregateSchemaFactory = "schema.MapAttribute{"
			typeName = "map"

			if isTopLevelAttribute {
				fprintf(e.StructWriter, "types.Map")
			}

			fwPlanModifierPackage = "mapplanmodifier"
			fwPlanModifierType = "Map"
			fwValidatorsPackage = "mapvalidator"
			fwValidatorType = "Map"

		case schema.TypeSet:
			aggregateSchemaFactory = "schema.SetAttribute{"
			typeName = "set"

			if isTopLevelAttribute {
				fprintf(e.StructWriter, "types.Set")
			}

			fwPlanModifierPackage = "setplanmodifier"
			fwPlanModifierType = "Set"
			fwValidatorsPackage = "setvalidator"
			fwValidatorType = "Set"
		}

		switch v := property.Elem.(type) {
		case *schema.Schema:
			var elementType string

			switch v := v.Type; v {
			case schema.TypeBool:
				elementType = "types.BoolType"

			case schema.TypeFloat:
				elementType = "types.Float64Type"

			case schema.TypeInt:
				elementType = "types.Int64Type"

			case schema.TypeString:
				elementType = "types.StringType"
				// Special handling for 'tags' and 'tags_all'.
				if typeName == "map" && isTopLevelAttribute {
					if attributeName == "tags" {
						e.HasTopLevelTagsMap = true
						if property.Optional {
							fprintf(e.SchemaWriter, "// TODO tftags.TagsAttribute()\n")
						} else if property.Computed {
							fprintf(e.SchemaWriter, "// TODO tftags.TagsAttributeComputedOnly()\n")
						}
					} else if attributeName == "tags_all" {
						e.HasTopLevelTagsAllMap = true
						fprintf(e.SchemaWriter, "// TODO tftags.TagsAttributeComputedOnly()\n")
					}
				}

			default:
				return unsupportedTypeError(path, fmt.Sprintf("(Attribute) %s of %s", typeName, v.String()))
			}

			fprintf(e.SchemaWriter, "%s\n", aggregateSchemaFactory)
			fprintf(e.SchemaWriter, "ElementType:%s,\n", elementType)

		case *schema.Resource:
			// We get here for Computed-only nested blocks or when ConfigMode is SchemaConfigModeBlock.
			fprintf(e.SchemaWriter, "%s\n", aggregateSchemaFactory)
			fprintf(e.SchemaWriter, "ElementType:")

			if err := e.emitComputedOnlyBlock(path, v.Schema); err != nil {
				return err
			}

			fprintf(e.SchemaWriter, ",\n")

		default:
			return unsupportedTypeError(path, fmt.Sprintf("(Attribute) %s of %T", typeName, v))
		}

	default:
		return unsupportedTypeError(path, v.String())
	}

	if property.Required {
		fprintf(e.SchemaWriter, "Required:true,\n")
	}

	if property.Optional {
		fprintf(e.SchemaWriter, "Optional:true,\n")
	}

	if def := property.Default; def != nil {
		property.Computed = true
	}

	if property.Computed {
		fprintf(e.SchemaWriter, "Computed:true,\n")
	}

	if property.Sensitive {
		fprintf(e.SchemaWriter, "Sensitive:true,\n")
	}

	if description := property.Description; description != "" {
		fprintf(e.SchemaWriter, "Description:%q,\n", description)
	}

	if deprecationMessage := property.Deprecated; deprecationMessage != "" {
		fprintf(e.SchemaWriter, "DeprecationMessage:%q,\n", deprecationMessage)
	}

	if maxItems, minItems := property.MaxItems, property.MinItems; maxItems > 0 || minItems > 0 && fwValidatorsPackage != "" && fwValidatorType != "" {
		e.FrameworkValidatorsPackages = append(e.FrameworkValidatorsPackages, fwValidatorsPackage)

		fprintf(e.SchemaWriter, "Validators:[]validator.%s{\n", fwValidatorType)
		if minItems > 0 {
			fprintf(e.SchemaWriter, "%s.SizeAtLeast(%d),\n", fwValidatorsPackage, minItems)
		}
		if maxItems > 0 {
			fprintf(e.SchemaWriter, "%s.SizeAtMost(%d),\n", fwValidatorsPackage, maxItems)
		}
		fprintf(e.SchemaWriter, "},\n")
	}

	if attributeName == "id" && isTopLevelAttribute && !e.IsDataSource {
		planModifiers = append(planModifiers, fmt.Sprintf("%s.UseStateForUnknown()", fwPlanModifierPackage))
		e.FrameworkPlanModifierPackages = append(e.FrameworkPlanModifierPackages, fwPlanModifierPackage)
	}

	if property.ForceNew {
		planModifiers = append(planModifiers, fmt.Sprintf("%s.RequiresReplace()", fwPlanModifierPackage))
		e.FrameworkPlanModifierPackages = append(e.FrameworkPlanModifierPackages, fwPlanModifierPackage)
	}

	if def := property.Default; def != nil {
		switch v := def.(type) {
		case bool:
			e.GoImports = append(e.GoImports, goImport{
				Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault",
			})
			defaultSpec = fmt.Sprintf("booldefault.StaticBool(%t)", v)
		case int:
			e.GoImports = append(e.GoImports, goImport{
				Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default",
			})
			defaultSpec = fmt.Sprintf("int64default.StaticInt64(%d)", v)
		case float64:
			e.GoImports = append(e.GoImports, goImport{
				Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default",
			})
			defaultSpec = fmt.Sprintf("float64default.StaticFloat64(%f)", v)
		case string:
			e.GoImports = append(e.GoImports, goImport{
				Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault",
			})
			defaultSpec = fmt.Sprintf("stringdefault.StaticString(%q)", v)
		default:
			fprintf(e.SchemaWriter, "// TODO Default: %#[1]v (%[1]T),\n", def)
		}
	}

	if len(planModifiers) > 0 {
		fprintf(e.SchemaWriter, "PlanModifiers:[]planmodifier.%s{\n", fwPlanModifierType)
		for _, planModifier := range planModifiers {
			fprintf(e.SchemaWriter, "%s,\n", planModifier)
		}
		fprintf(e.SchemaWriter, "},\n")
	}

	if defaultSpec != "" {
		fprintf(e.SchemaWriter, "Default: %s,\n", defaultSpec)
	}

	// Features that we can't (yet) migrate:

	if property.ValidateFunc != nil || property.ValidateDiagFunc != nil {
		fprintf(e.SchemaWriter, "// TODO Validate,\n")
	}

	fprintf(e.SchemaWriter, "}")

	return nil
}

// emitBlockProperty generates the Plugin Framework code for a Plugin SDK Block's property
// and emits the generated code to the emitter's Writer.
func (e *emitter) emitBlockProperty(path []string, property *schema.Schema) error {
	var planModifiers []string
	var fwPlanModifierPackage, fwPlanModifierType, fwValidatorsPackage, fwValidatorType string

	// At this point we are emitting code for the values of a schema.Block or Schema's Blocks (map[string]schema.Block).
	switch v := property.Type; v {
	//
	// Complex types.
	//
	case schema.TypeList:
		switch v := property.Elem.(type) {
		case *schema.Resource:
			fwPlanModifierPackage = "listplanmodifier"
			fwPlanModifierType = "List"
			fwValidatorsPackage = "listvalidator"
			fwValidatorType = "List"

			fprintf(e.SchemaWriter, "schema.ListNestedBlock{\n")
			fprintf(e.SchemaWriter, "NestedObject:schema.NestedBlockObject{\n")

			err := e.emitAttributesAndBlocks(path, v.Schema)

			if err != nil {
				return err
			}

			fprintf(e.SchemaWriter, "},\n")

		default:
			return unsupportedTypeError(path, fmt.Sprintf("(Block) list of %T", v))
		}

	case schema.TypeSet:
		switch v := property.Elem.(type) {
		case *schema.Resource:
			fwPlanModifierPackage = "setplanmodifier"
			fwPlanModifierType = "Set"
			fwValidatorsPackage = "setvalidator"
			fwValidatorType = "Set"

			fprintf(e.SchemaWriter, "schema.SetNestedBlock{\n")
			fprintf(e.SchemaWriter, "NestedObject:schema.NestedBlockObject{\n")

			err := e.emitAttributesAndBlocks(path, v.Schema)

			if err != nil {
				return err
			}

			fprintf(e.SchemaWriter, "},\n")

		default:
			return unsupportedTypeError(path, fmt.Sprintf("(Block) set of %T", v))
		}

	default:
		return unsupportedTypeError(path, v.String())
	}

	// Compatibility hacks.
	// See Schema::coreConfigSchemaBlock.
	if property.Required && property.MinItems == 0 {
		property.MinItems = 1
	}
	if property.Optional && property.MinItems > 0 {
		property.MinItems = 0
	}
	if property.Computed && !property.Optional {
		property.MaxItems = 0
		property.MinItems = 0
	}

	if description := property.Description; description != "" {
		fprintf(e.SchemaWriter, "Description:%q,\n", description)
	}

	if deprecationMessage := property.Deprecated; deprecationMessage != "" {
		fprintf(e.SchemaWriter, "DeprecationMessage:%q,\n", deprecationMessage)
	}

	if maxItems, minItems := property.MaxItems, property.MinItems; maxItems > 0 || minItems > 0 && fwValidatorsPackage != "" && fwValidatorType != "" {
		e.FrameworkValidatorsPackages = append(e.FrameworkValidatorsPackages, fwValidatorsPackage)

		fprintf(e.SchemaWriter, "Validators:[]validator.%s{\n", fwValidatorType)
		if minItems > 0 {
			fprintf(e.SchemaWriter, "%s.SizeAtLeast(%d),\n", fwValidatorsPackage, minItems)
		}
		if maxItems > 0 {
			fprintf(e.SchemaWriter, "%s.SizeAtMost(%d),\n", fwValidatorsPackage, maxItems)
		}
		fprintf(e.SchemaWriter, "},\n")
	}

	if property.ForceNew {
		planModifiers = append(planModifiers, fmt.Sprintf("%s.RequiresReplace()", fwPlanModifierPackage))
		e.FrameworkPlanModifierPackages = append(e.FrameworkPlanModifierPackages, fwPlanModifierPackage)
	}

	if len(planModifiers) > 0 {
		fprintf(e.SchemaWriter, "PlanModifiers:[]planmodifier.%s{\n", fwPlanModifierType)
		for _, planModifier := range planModifiers {
			fprintf(e.SchemaWriter, "%s,\n", planModifier)
		}
		fprintf(e.SchemaWriter, "},\n")
	}

	if def := property.Default; def != nil {
		e.warnf("Block %s has non-nil Default: %v", strings.Join(path, "/"), def)
	}

	fprintf(e.SchemaWriter, "}")

	return nil
}

// emitComputedOnlyBlock generates the Plugin Framework code for a Plugin SDK Computed-only nested block
// and emits the generated code to the emitter's Writer.
// See https://github.com/hashicorp/terraform-plugin-sdk/blob/6ffc92796f0716c07502e4d36aaafa5fd85e94cf/internal/configs/configschema/implied_type.go#L12.
// Property names are sorted prior to code generation to reduce diffs.
func (e *emitter) emitComputedOnlyBlock(path []string, schema map[string]*schema.Schema) error {
	names := make([]string, 0)
	for name := range schema {
		names = append(names, name)
	}
	sort.Strings(names)

	fprintf(e.SchemaWriter, "types.ObjectType{\n")

	emittedFieldName := false
	for _, name := range names {
		property := schema[name]

		if !emittedFieldName {
			fprintf(e.SchemaWriter, "AttrTypes: map[string]attr.Type{\n")
			emittedFieldName = true
			e.ImportFrameworkAttr = true
		}

		fprintf(e.SchemaWriter, "%q:", name)

		err := e.emitComputedOnlyBlockProperty(append(path, name), property)

		if err != nil {
			return err
		}
	}
	if emittedFieldName {
		fprintf(e.SchemaWriter, "},\n")
	}

	fprintf(e.SchemaWriter, "}")

	return nil
}

// emitComputedOnlyBlockProperty generates the Plugin Framework code for a Plugin SDK Computed-only nested block's property
// and emits the generated code to the emitter's Writer.
// See https://github.com/hashicorp/terraform-plugin-sdk/blob/6ffc92796f0716c07502e4d36aaafa5fd85e94cf/internal/configs/configschema/implied_type.go#L12.
func (e *emitter) emitComputedOnlyBlockProperty(path []string, property *schema.Schema) error {
	// At this point we are emitting code for the values of a types.ObjectType's AttrMap (map[string]attr.Type).
	switch v := property.Type; v {
	//
	// Primitive types.
	//
	case schema.TypeBool:
		fprintf(e.SchemaWriter, "types.BoolType,\n")

	case schema.TypeFloat:
		fprintf(e.SchemaWriter, "types.Float64Type,\n")

	case schema.TypeInt:
		fprintf(e.SchemaWriter, "types.Int64Type,\n")

	case schema.TypeString:
		fprintf(e.SchemaWriter, "types.StringType,\n")

	//
	// Complex types.
	//
	case schema.TypeList, schema.TypeMap, schema.TypeSet:
		var aggregateType, typeName string

		switch v {
		case schema.TypeList:
			aggregateType = "types.ListType"
			typeName = "list"
		case schema.TypeMap:
			aggregateType = "types.MapType"
			typeName = "map"
		case schema.TypeSet:
			aggregateType = "types.SetType"
			typeName = "set"
		}

		switch v := property.Elem.(type) {
		case *schema.Schema:
			var elementType string

			switch v := v.Type; v {
			case schema.TypeBool:
				elementType = "types.BoolType"

			case schema.TypeFloat:
				elementType = "types.Float64Type"

			case schema.TypeInt:
				elementType = "types.Int64Type"

			case schema.TypeString:
				elementType = "types.StringType"

			default:
				return unsupportedTypeError(path, fmt.Sprintf("(ComputedOnlyBlockProperty) %s of %s", typeName, v.String()))
			}

			fprintf(e.SchemaWriter, "%s{ElemType:%s},\n", aggregateType, elementType)

		case *schema.Resource:
			// We get here for Computed-only nested blocks or when ConfigMode is SchemaConfigModeBlock.
			fprintf(e.SchemaWriter, "%s{ElemType:", aggregateType)

			if err := e.emitComputedOnlyBlock(path, v.Schema); err != nil {
				return err
			}

			fprintf(e.SchemaWriter, "},\n")

		default:
			return unsupportedTypeError(path, fmt.Sprintf("(ComputedOnlyBlockProperty) %s of %T", typeName, v))
		}

	default:
		return unsupportedTypeError(path, v.String())
	}

	return nil
}

// warnf emits a formatted warning message to the UI.
func (e *emitter) warnf(format string, a ...interface{}) {
	e.Generator.Warnf(format, a...)
}

// fprintf writes a formatted string to a Writer.
func fprintf(w io.Writer, format string, a ...interface{}) (int, error) {
	return io.WriteString(w, fmt.Sprintf(format, a...))
}

// isAttribute returns whether or not the specified property should be emitted as an Attribute (vs. a Block).
// See https://github.com/hashicorp/terraform-plugin-sdk/blob/6ffc92796f0716c07502e4d36aaafa5fd85e94cf/helper/schema/core_schema.go#L57.
func isAttribute(property *schema.Schema) bool {
	if property.Elem == nil {
		return true
	}

	if property.Type == schema.TypeMap {
		return true
	}

	switch property.ConfigMode {
	case schema.SchemaConfigModeAttr:
		return true

	case schema.SchemaConfigModeBlock:
		return false

	default:
		if property.Computed && !property.Optional {
			// Computed-only schemas are always handled as attributes because they never appear in configuration.
			return true
		}

		switch property.Elem.(type) {
		case *schema.Schema:
			return true
		}
	}

	return false
}

func unsupportedTypeError(path []string, typ string) error {
	return fmt.Errorf("%s is of unsupported type: %s", strings.Join(path, "/"), typ)
}

type templateData struct {
	DefaultCreateTimeout          int64
	DefaultReadTimeout            int64
	DefaultUpdateTimeout          int64
	DefaultDeleteTimeout          int64
	EmitResourceImportState       bool
	EmitResourceModifyPlan        bool
	EmitResourceUpdateSkeleton    bool
	FrameworkPlanModifierPackages []string
	FrameworkValidatorsPackages   []string
	GoImports                     []goImport
	HasTimeouts                   bool
	ImportFrameworkAttr           bool
	ImportProviderFrameworkTypes  bool
	Name                          string // e.g. Instance
	PackageName                   string // e.g. ec2
	Schema                        string
	Struct                        string
	TFTypeName                    string // e.g. aws_instance
}

//go:embed datasource.gtpl
var datasourceImpl string

//go:embed resource.gtpl
var resourceImpl string

type goImport struct {
	Path  string
	Alias string
}
