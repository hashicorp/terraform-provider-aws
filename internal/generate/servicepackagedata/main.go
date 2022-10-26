//go:build generate
// +build generate

package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"go/format"
	"os"
	"text/template"

	"github.com/mitchellh/cli"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\tmain.go [flags] [<generated-file>]\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	filename := `service_package_data_gen.go`
	if args := flag.Args(); len(args) > 0 {
		filename = args[0]
	}

	ui := &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}
	generator := &Generator{
		UI: ui,
	}
	templateData := &TemplateData{
		PackageName: os.Getenv("GOPACKAGE"),
	}

	if err := generator.ApplyAndWriteTemplate(filename, tmpl, templateData); err != nil {
		ui.Error(fmt.Sprintf("error generating %s service package data: %s", templateData.PackageName, err.Error()))
		os.Exit(1)
	}
}

type Generator struct {
	UI cli.Ui
}

func (g *Generator) ApplyAndWriteTemplate(filename, templateBody string, templateData *TemplateData) error {
	tmpl, err := template.New("servicepackagedata").Parse(templateBody)

	if err != nil {
		return fmt.Errorf("parsing function template: %w", err)
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, templateData)

	if err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	generatedFileContents, err := format.Source(buffer.Bytes())

	if err != nil {
		g.Infof("%s", buffer.String())
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

func (g *Generator) Infof(format string, a ...interface{}) {
	g.UI.Info(fmt.Sprintf(format, a...))
}

type TemplateData struct {
	PackageName string
}

//go:embed file.tmpl
var tmpl string
