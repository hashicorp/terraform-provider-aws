//go:build ignore
// +build ignore

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
	fmt.Fprintf(os.Stderr, "\tmain.go [flags] -resource <TF-resource-type> <generated-schema-file>\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
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
	err := e.emitSchema(schema)

	if err != nil {
		return err
	}

	return nil
}

// emitSchema generates the Plugin Framework code for a Plugin SDK schema and emits the generated code to the emitter's Writer.
// A schema is a map of property names to Attributes.
// Property names are sorted prior to code generation to reduce diffs.
func (e emitter) emitSchema(schema map[string]*schema.Schema) error {
	names := make([]string, 0)
	for name := range schema {
		names = append(names, name)
	}
	sort.Strings(names)

	e.printf("map[string]tfsdk.Attribute{\n")
	for _, name := range names {
		e.printf("%q:", name)

		err := e.emitAttribute(schema[name])

		if err != nil {
			return err
		}

		e.printf(",\n")
	}
	e.printf("}")

	return nil
}

// emitAttribute generates the Plugin Framework code for a Plugin SDK property's Attributes and emits the generated code to the emitter's Writer.
func (e emitter) emitAttribute(property *schema.Schema) error {
	e.printf("{\n")
	if description := property.Description; description != "" {
		e.printf("Description:%q,\n", description)
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

type templateData struct {
	RootSchema string
}

var schemaTemplateBody = `
// Code generated by generates/schemamigrate/main.go; DO NOT EDIT.

var (
	schema = {{ .RootSchema }}
)
`
