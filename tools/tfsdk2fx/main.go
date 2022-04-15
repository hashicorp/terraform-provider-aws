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
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	"github.com/mitchellh/cli"
)

var (
	tfResourceType = flag.String("resource", "", "Terraform resource type")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\ttfsdk2fx -resource <TF-resource-type> <generated-schema-file>\n\n")
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 || *tfResourceType == "" {
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

	p := provider.Provider()

	if *tfResourceType == "provider" {
		migrator.ProviderSchema = p.Schema
	}

	if err := migrator.migrate(outputFilename); err != nil {
		ui.Error(fmt.Sprintf("error migrating Terraform %s schema: %s", *tfResourceType, err))
		os.Exit(1)
	}
}

type migrator struct {
	ProviderSchema map[string]*schema.Schema
	Ui             cli.Ui
}

// migrate generates an identical schema into the specified output file.
func (m *migrator) migrate(outputFilename string) error {
	m.infof("generating schema into %[1]q", outputFilename)

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
		Ui:     m.Ui,
		Writer: &sb,
	}

	var err error

	if len(m.ProviderSchema) > 0 {
		err = emitter.emitSchemaForProvider(m.ProviderSchema)
	}

	if err != nil {
		return nil, fmt.Errorf("emitting schema code: %w", err)
	}

	rootSchema := sb.String()
	templateData := &templateData{
		RootSchema: rootSchema,
	}

	return templateData, nil
}

func (m *migrator) infof(format string, a ...interface{}) {
	m.Ui.Info(fmt.Sprintf(format, a...))
}

type emitter struct {
	Ui     cli.Ui
	Writer io.Writer
}

// emitSchemaForProvider generates the Plugin Framework code for a Plugin SDK Provider schema and emits the generated code to the emitter's Writer.
func (e emitter) emitSchemaForProvider(schema map[string]*schema.Schema) error {
	e.printf("tfsdk.Schema{\n")

	err := e.emitAttributesAndBlocks(nil, schema)

	if err != nil {
		return err
	}

	e.printf("}")

	return nil
}

// emitSchemaForResource generates the Plugin Framework code for a Plugin SDK Resource and emits the generated code to the emitter's Writer.
func (e emitter) emitSchemaForResource(resource *schema.Resource) error {
	e.printf("tfsdk.Schema{\n")

	err := e.emitAttributesAndBlocks(nil, resource.Schema)

	if err != nil {
		return err
	}

	if version := resource.SchemaVersion; version > 0 {
		e.printf("Version:%q,\n", version)
	}

	if description := resource.Description; description != "" {
		e.printf("Description:%q,\n", description)
	}

	if deprecationMessage := resource.DeprecationMessage; deprecationMessage != "" {
		e.printf("DeprecationMessage:%q,\n", deprecationMessage)
	}

	e.printf("}")

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

	e.printf("Attributes: map[string]tfsdk.Attribute{\n")
	for _, name := range names {
		property := schema[name]

		if !isAttribute(property) {
			continue
		}

		e.printf("%q:", name)

		err := e.emitAttribute(append(path, name), property)

		if err != nil {
			return err
		}

		e.printf(",\n")
	}
	e.printf("},\n")

	e.printf("Blocks: map[string]tfsdk.Block{\n")
	for _, name := range names {
		property := schema[name]

		if isAttribute(property) {
			continue
		}

		e.printf("%q:", name)

		err := e.emitBlock(append(path, name), property)

		if err != nil {
			return err
		}

		e.printf(",\n")
	}
	e.printf("},\n")

	return nil
}

// emitAttribute generates the Plugin Framework code for a Plugin SDK Attribute property
// and emits the generated code to the emitter's Writer.
func (e emitter) emitAttribute(path []string, property *schema.Schema) error {
	e.printf("{\n")

	switch v := property.Type; v {
	//
	// Primitive types.
	//
	case schema.TypeBool:
		e.printf("Type:types.BoolType,\n")

	case schema.TypeFloat:
		e.printf("Type:types.Float64Type,\n")

	case schema.TypeInt:
		e.printf("Type:types.Int64Type,\n")

	case schema.TypeString:
		e.printf("Type:types.StringType,\n")

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
				return unsupportedTypeError(path, fmt.Sprintf("list of %s", v.String()))
			}

			e.printf("Type:types.ListType{ElemType:%s},\n", elementType)

		default:
			return unsupportedTypeError(path, fmt.Sprintf("list of %T", v))
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
				return unsupportedTypeError(path, fmt.Sprintf("map of %s", v.String()))
			}

			e.printf("Type:types.MapType{ElemType:%s},\n", elementType)

		default:
			return unsupportedTypeError(path, fmt.Sprintf("map of %T", v))
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
				return unsupportedTypeError(path, fmt.Sprintf("set of %s", v.String()))
			}

			e.printf("Type:types.SetType{ElemType:%s},\n", elementType)

		default:
			return unsupportedTypeError(path, fmt.Sprintf("set of %T", v))
		}

	default:
		return unsupportedTypeError(path, v.String())
	}

	if property.Required {
		e.printf("Required:true,\n")
	}

	if property.Optional {
		e.printf("Optional:true,\n")
	}

	if property.Computed {
		e.printf("Computed:true,\n")
	}

	if property.Sensitive {
		e.printf("Sensitive:true,\n")
	}

	if description := property.Description; description != "" {
		e.printf("Description:%q,\n", description)
	}

	if deprecationMessage := property.Deprecated; deprecationMessage != "" {
		e.printf("DeprecationMessage:%q,\n", deprecationMessage)
	}

	// Features that we can't (yet) migrate:

	if property.ForceNew {
		e.printf("// TODO ForceNew:true,\n")
	}

	e.printf("}")

	return nil
}

// emitBlock generates the Plugin Framework code for a Plugin SDK Block property
// and emits the generated code to the emitter's Writer.
func (e emitter) emitBlock(path []string, property *schema.Schema) error {
	e.printf("{\n")

	switch v := property.Type; v {
	//
	// Complex types.
	//
	case schema.TypeList:
		switch v := property.Elem.(type) {
		case *schema.Resource:
			e.printf("NestingMode:tfsdk.BlockNestingModeList,\n")

		default:
			return unsupportedTypeError(path, fmt.Sprintf("list of %T", v))
		}

	case schema.TypeSet:
		switch v := property.Elem.(type) {
		case *schema.Resource:
			e.printf("NestingMode:tfsdk.BlockNestingModeSet,\n")

		default:
			return unsupportedTypeError(path, fmt.Sprintf("set of %T", v))
		}

	default:
		return unsupportedTypeError(path, v.String())
	}

	if maxItems := property.MaxItems; maxItems > 1 {
		e.printf("MaxItems:%d,\n", maxItems)
	}

	if minItems := property.MinItems; minItems > 1 {
		e.printf("MinItems:%d,\n", minItems)
	}

	if description := property.Description; description != "" {
		e.printf("Description:%q,\n", description)
	}

	if deprecationMessage := property.Deprecated; deprecationMessage != "" {
		e.printf("DeprecationMessage:%q,\n", deprecationMessage)
	}

	e.printf("}")

	return nil
}

// printf emits a formatted string to the underlying writer.
func (e emitter) printf(format string, a ...interface{}) (int, error) {
	return fprintf(e.Writer, format, a...)
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
	RootSchema string
}

var schemaTemplateBody = `
// Code generated by tools/tfsdk2fx/main.go; DO NOT EDIT.

var (
	schema = {{ .RootSchema }}
)
`
