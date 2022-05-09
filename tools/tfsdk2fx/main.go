package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/tf5provider"
	"github.com/mitchellh/cli"
)

var (
	dataSourceType  = flag.String("data-source", "", "Data Source type")
	migrateProvider = flag.Bool("provider", false, "Migrate provider schema")
	resourceType    = flag.String("resource", "", "Resource type")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\ttfsdk2fx [-provider|-resource <resource-type>|-data-source <data-source-type>] <generated-file>\n\n")
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 || (*dataSourceType == "" && !*migrateProvider && *resourceType == "") {
		flag.Usage()
		os.Exit(2)
	}

	outputFilename := args[0]

	ui := &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}
	migrator := &migrator{
		Ui: ui,
	}

	p := tf5provider.Provider()

	if *migrateProvider {
		migrator.ProviderSchema = p.Schema
	} else if v := *dataSourceType; v != "" {
		resource, ok := p.DataSourcesMap[v]

		if !ok {
			ui.Error(fmt.Sprintf("data source type %s not found", v))
			os.Exit(2)
		}

		migrator.Resource = resource
	} else if v := *resourceType; v != "" {
		resource, ok := p.ResourcesMap[v]

		if !ok {
			ui.Error(fmt.Sprintf("resource type %s not found", v))
			os.Exit(2)
		}

		migrator.Resource = resource
	}

	if err := migrator.migrate(outputFilename); err != nil {
		ui.Error(fmt.Sprintf("error migrating Terraform %s schema: %s", *resourceType, err))
		os.Exit(1)
	}
}

type migrator struct {
	ProviderSchema map[string]*schema.Schema
	Resource       *schema.Resource
	Ui             cli.Ui
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

	err = m.applyTemplate(outputFilename, schemaTemplateBody, templateData)

	if err != nil {
		return err
	}

	return nil
}

func (m *migrator) applyTemplate(filename, templateBody string, templateData *templateData) error {
	tmpl, err := template.New("schema").Parse(templateBody)

	if err != nil {
		return fmt.Errorf("parsing schema template: %w", err)
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, templateData)

	if err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	generatedFileContents, err := format.Source(buffer.Bytes())

	if err != nil {
		m.infof("%s", buffer.String())
		return fmt.Errorf("formatting generated source code: %w", err)
	}

	f, err := os.Create(filename)

	if err != nil {
		return fmt.Errorf("creating file (%s): %w", filename, err)
	}

	defer f.Close()

	_, err = f.Write(generatedFileContents)

	if err != nil {
		return fmt.Errorf("writing to file (%s): %w", filename, err)
	}

	return nil
}

func (m *migrator) generateTemplateData() (*templateData, error) {
	sb := strings.Builder{}
	emitter := &emitter{
		Ui:           m.Ui,
		SchemaWriter: &sb,
	}

	var err error

	if len(m.ProviderSchema) > 0 {
		err = emitter.emitSchemaForProvider(m.ProviderSchema)
	} else if m.Resource != nil {
		err = emitter.emitSchemaForResource(m.Resource)
	}

	if err != nil {
		return nil, fmt.Errorf("emitting schema code: %w", err)
	}

	schema := sb.String()
	templateData := &templateData{
		Schema: schema,
	}

	return templateData, nil
}

func (m *migrator) infof(format string, a ...interface{}) {
	m.Ui.Info(fmt.Sprintf(format, a...))
}

type emitter struct {
	Ui           cli.Ui
	SchemaWriter io.Writer
}

// emitSchemaForProvider generates the Plugin Framework code for a Plugin SDK Provider schema and emits the generated code to the emitter's Writer.
func (e emitter) emitSchemaForProvider(schema map[string]*schema.Schema) error {
	fprintf(e.SchemaWriter, "tfsdk.Schema{\n")

	err := e.emitAttributesAndBlocks(nil, schema)

	if err != nil {
		return err
	}

	fprintf(e.SchemaWriter, "}")

	return nil
}

// emitSchemaForResource generates the Plugin Framework code for a Plugin SDK Resource and emits the generated code to the emitter's Writer.
func (e emitter) emitSchemaForResource(resource *schema.Resource) error {
	if _, ok := resource.Schema["id"]; ok {
		e.warnf("Explicit `id` attribute defined")
	} else {
		resource.Schema["id"] = &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		}
	}

	fprintf(e.SchemaWriter, "tfsdk.Schema{\n")

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

// emitAttributesAndBlocks generates the Plugin Framework code for a set of Plugin SDK Attribute and Block properties
// and emits the generated code to the emitter's Writer.
// Property names are sorted prior to code generation to reduce diffs.
func (e emitter) emitAttributesAndBlocks(path []string, schema map[string]*schema.Schema) error {
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
			fprintf(e.SchemaWriter, "Attributes: map[string]tfsdk.Attribute{\n")
			emittedFieldName = true
		}

		fprintf(e.SchemaWriter, "%q:", name)

		err := e.emitAttribute(append(path, name), property)

		if err != nil {
			return err
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
			fprintf(e.SchemaWriter, "Blocks: map[string]tfsdk.Block{\n")
			emittedFieldName = true
		}

		fprintf(e.SchemaWriter, "%q:", name)

		err := e.emitBlock(append(path, name), property)

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

// emitAttribute generates the Plugin Framework code for a Plugin SDK Attribute property
// and emits the generated code to the emitter's Writer.
func (e emitter) emitAttribute(path []string, property *schema.Schema) error {
	fprintf(e.SchemaWriter, "{\n")

	switch v := property.Type; v {
	//
	// Primitive types.
	//
	case schema.TypeBool:
		fprintf(e.SchemaWriter, "Type:types.BoolType,\n")

	case schema.TypeFloat:
		fprintf(e.SchemaWriter, "Type:types.Float64Type,\n")

	case schema.TypeInt:
		fprintf(e.SchemaWriter, "Type:types.Int64Type,\n")

	case schema.TypeString:
		fprintf(e.SchemaWriter, "Type:types.StringType,\n")

	//
	// Complex types.
	//
	case schema.TypeList:
		switch v := property.Elem.(type) {
		case *schema.Schema:
			//
			// List of primitives.
			//
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
				return unsupportedTypeError(path, fmt.Sprintf("(Attribute) list of %s", v.String()))
			}

			fprintf(e.SchemaWriter, "Type:types.ListType{ElemType:%s},\n", elementType)

		case *schema.Resource:
			// We get here for Computed-only nested blocks or when ConfigMode is SchemaConfigModeBlock.
			return unsupportedTypeError(path, fmt.Sprintf("(Attribute) list of %T", v))

		default:
			return unsupportedTypeError(path, fmt.Sprintf("(Attribute) list of %T", v))
		}

	case schema.TypeMap:
		switch v := property.Elem.(type) {
		case *schema.Schema:
			//
			// Map of primitives.
			//
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
				return unsupportedTypeError(path, fmt.Sprintf("(Attribute) map of %s", v.String()))
			}

			fprintf(e.SchemaWriter, "Type:types.MapType{ElemType:%s},\n", elementType)

		case *schema.Resource:
			// We get here for Computed-only nested blocks or when ConfigMode is SchemaConfigModeBlock.
			return unsupportedTypeError(path, fmt.Sprintf("(Attribute) list of %T", v))

		default:
			return unsupportedTypeError(path, fmt.Sprintf("(Attribute) map of %T", v))
		}

	case schema.TypeSet:
		switch v := property.Elem.(type) {
		case *schema.Schema:
			//
			// Set of primitives.
			//
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
				return unsupportedTypeError(path, fmt.Sprintf("(Attribute) set of %s", v.String()))
			}

			fprintf(e.SchemaWriter, "Type:types.SetType{ElemType:%s},\n", elementType)

		case *schema.Resource:
			// We get here for Computed-only nested blocks or when ConfigMode is SchemaConfigModeBlock.
			return unsupportedTypeError(path, fmt.Sprintf("(Attribute) list of %T", v))

		default:
			return unsupportedTypeError(path, fmt.Sprintf("(Attribute) set of %T", v))
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

	// Features that we can't (yet) migrate:

	if property.ForceNew {
		fprintf(e.SchemaWriter, "// TODO ForceNew:true,\n")
	}

	if def := property.Default; def != nil {
		switch v := def.(type) {
		case bool:
			if v {
				fprintf(e.SchemaWriter, "// TODO Default:%#v,\n", def)
			} else {
				e.warnf("Attribute %s has spurious Default: %#v", strings.Join(path, "/"), def)
			}
		case int:
			if v != 0 {
				fprintf(e.SchemaWriter, "// TODO Default:%#v,\n", def)
			} else {
				e.warnf("Attribute %s has spurious Default: %#v", strings.Join(path, "/"), def)
			}
		case float64:
			if v != 0 {
				fprintf(e.SchemaWriter, "// TODO Default:%#v,\n", def)
			} else {
				e.warnf("Attribute %s has spurious Default: %#v", strings.Join(path, "/"), def)
			}
		case string:
			if v != "" {
				fprintf(e.SchemaWriter, "// TODO Default:%#v,\n", def)
			} else {
				e.warnf("Attribute %s has spurious Default: %#v", strings.Join(path, "/"), def)
			}
		default:
		}
	}

	fprintf(e.SchemaWriter, "}")

	return nil
}

// emitBlock generates the Plugin Framework code for a Plugin SDK Block property
// and emits the generated code to the emitter's Writer.
func (e emitter) emitBlock(path []string, property *schema.Schema) error {
	fprintf(e.SchemaWriter, "{\n")

	switch v := property.Type; v {
	//
	// Complex types.
	//
	case schema.TypeList:
		switch v := property.Elem.(type) {
		case *schema.Resource:
			err := e.emitAttributesAndBlocks(path, v.Schema)

			if err != nil {
				return err
			}

			fprintf(e.SchemaWriter, "NestingMode:tfsdk.BlockNestingModeList,\n")

		default:
			return unsupportedTypeError(path, fmt.Sprintf("(Block) list of %T", v))
		}

	case schema.TypeSet:
		switch v := property.Elem.(type) {
		case *schema.Resource:
			err := e.emitAttributesAndBlocks(path, v.Schema)

			if err != nil {
				return err
			}

			fprintf(e.SchemaWriter, "NestingMode:tfsdk.BlockNestingModeSet,\n")

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

	if maxItems := property.MaxItems; maxItems > 0 {
		fprintf(e.SchemaWriter, "MaxItems:%d,\n", maxItems)
	}

	if minItems := property.MinItems; minItems > 0 {
		fprintf(e.SchemaWriter, "MinItems:%d,\n", minItems)
	}

	if description := property.Description; description != "" {
		fprintf(e.SchemaWriter, "Description:%q,\n", description)
	}

	if deprecationMessage := property.Deprecated; deprecationMessage != "" {
		fprintf(e.SchemaWriter, "DeprecationMessage:%q,\n", deprecationMessage)
	}

	if def := property.Default; def != nil {
		e.warnf("Block %s has non-nil Default: %v", strings.Join(path, "/"), def)
	}

	fprintf(e.SchemaWriter, "}")

	return nil
}

// warnf emits a formatted warning message to the UI.
func (e emitter) warnf(format string, a ...interface{}) {
	e.Ui.Warn(fmt.Sprintf(format, a...))
}

// fprintf writes a formatted string to a Writer.
func fprintf(w io.Writer, format string, a ...interface{}) (int, error) {
	return io.WriteString(w, fmt.Sprintf(format, a...))
}

// isAttribute returns whether or not the specified property should be emitted as an Attribute.
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
	Schema string
}

var schemaTemplateBody = `
// Code generated by tools/tfsdk2fx/main.go; DO NOT EDIT.

schema := {{ .Schema }}
`
