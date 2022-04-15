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
	migrator := &schemaMigrator{
		Ui: ui,
	}

	p := provider.Provider()

	if *tfResourceType == "provider" {
		migrator.SDKSchema = p.Schema
	}

	if err := migrator.migrate(outputFilename); err != nil {
		ui.Error(fmt.Sprintf("error migrating Terraform %s schema: %s", *tfResourceType, err))
		os.Exit(1)
	}
}

type schemaMigrator struct {
	SDKSchema map[string]*schema.Schema
	Ui        cli.Ui
}

// migrate generates an identical schema into the specified output file.
func (m *schemaMigrator) migrate(outputFilename string) error {
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

func (m *schemaMigrator) applyTemplate(filename, templateBody string, templateData *templateData) error {
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

func (m *schemaMigrator) generateTemplateData() (*templateData, error) {
	sb := strings.Builder{}
	emitter := &emitter{
		Ui:     m.Ui,
		Writer: &sb,
	}

	err := emitter.emitRootSchema(m.SDKSchema)

	if err != nil {
		return nil, fmt.Errorf("emitting schema code: %w", err)
	}

	rootSchema := sb.String()
	templateData := &templateData{
		RootSchema: rootSchema,
	}

	return templateData, nil
}

func (m *schemaMigrator) infof(format string, a ...interface{}) {
	m.Ui.Info(fmt.Sprintf(format, a...))
}

type emitter struct {
	Ui     cli.Ui
	Writer io.Writer
}

// emitRootSchema generates the Plugin Framework code for a Plugin SDK root schema and emits the generated code to the emitter's Writer.
// The root schema is the map of root property names to Attributes.
func (e emitter) emitRootSchema(schema map[string]*schema.Schema) error {
	err := e.emitSchema(nil, schema)

	if err != nil {
		return err
	}

	return nil
}

// emitSchema generates the Plugin Framework code for a Plugin SDK schema and emits the generated code to the emitter's Writer.
// A schema is a map of property names to Attributes.
// Property names are sorted prior to code generation to reduce diffs.
func (e emitter) emitSchema(path []string, schema map[string]*schema.Schema) error {
	names := make([]string, 0)
	for name := range schema {
		names = append(names, name)
	}
	sort.Strings(names)

	e.printf("Attributes: map[string]tfsdk.Attribute{\n")
	for _, name := range names {
		err := e.emitAttribute(name, append(path, name), schema[name])

		if err != nil {
			return err
		}
	}
	e.printf("},\n")

	e.printf("Blocks: map[string]tfsdk.Block{\n")
	for _, name := range names {
		err := e.emitBlock(name, append(path, name), schema[name])

		if err != nil {
			return err
		}
	}
	e.printf("},\n")

	return nil
}

// emitAttribute generates the Plugin Framework code for a Plugin SDK property's Attribute and emits the generated code to the emitter's Writer.
func (e emitter) emitAttribute(name string, path []string, property *schema.Schema) error {
	if !e.isAttribute(property) {
		return nil
	}

	e.printf("%q:{\n", name)

	if description := property.Description; description != "" {
		e.printf("Description:%q,\n", description)
	}

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
		case *schema.Resource:
			//
			// Emitted as a Block.
			//
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
		case *schema.Resource:
			//
			// Emitted as a Block.
			//
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

	e.printf("},\n")

	return nil
}

// emitBlock generates the Plugin Framework code for a Plugin SDK property's Block and emits the generated code to the emitter's Writer.
func (e emitter) emitBlock(name string, path []string, property *schema.Schema) error {
	if e.isAttribute(property) {
		return nil
	}

	e.printf("%q:{\n", name)

	if description := property.Description; description != "" {
		e.printf("Description:%q,\n", description)
	}

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
		case *schema.Resource:
			//
			// Emitted as a Block.
			//
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
		case *schema.Resource:
			//
			// Emitted as a Block.
			//
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

	e.printf("},\n")

	return nil
}

// isAttribute returns whether or not the specified property should be emitted as an Attribute.
func (e emitter) isAttribute(property *schema.Schema) bool {
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

func unsupportedTypeError(path []string, typ string) error {
	return fmt.Errorf("%s is of unsupported type: %s", strings.Join(path, "/"), typ)
}

type templateData struct {
	RootSchema string
}

var schemaTemplateBody = `
// Code generated by generates/schemamigrate/main.go; DO NOT EDIT.

var (
	schema = tfsdk.Schema{
		{{ .RootSchema }}
	}
)
`
