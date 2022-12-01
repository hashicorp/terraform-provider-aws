package common

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"text/template"

	"github.com/mitchellh/cli"
)

type Generator struct {
	ui cli.Ui
}

func NewGenerator() *Generator {
	return &Generator{
		ui: &cli.BasicUi{
			Reader:      os.Stdin,
			Writer:      os.Stdout,
			ErrorWriter: os.Stderr,
		},
	}
}

func (g *Generator) ApplyAndWriteTemplate(filename, templateName, templateBody string, templateData any, formatter func([]byte) ([]byte, error)) error {
	tmpl, err := template.New(templateName).Parse(templateBody)

	if err != nil {
		return fmt.Errorf("parsing function template: %w", err)
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, templateData)

	if err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	var generatedFileContents []byte

	if formatter != nil {
		generatedFileContents, err = formatter(buffer.Bytes())

		if err != nil {
			g.Infof("%s", buffer.String())
			return fmt.Errorf("formatting generated source code: %w", err)
		}
	} else {
		generatedFileContents = buffer.Bytes()
	}

	f, err := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644) //nolint:gomnd

	if err != nil {
		return fmt.Errorf("opening file (%s): %w", filename, err)
	}

	defer f.Close()

	_, err = f.Write(generatedFileContents)

	if err != nil {
		return fmt.Errorf("writing to file (%s): %w", filename, err)
	}

	return nil
}

func (g *Generator) ApplyAndWriteGoTemplate(filename, templateName, templateBody string, templateData any) error {
	return g.ApplyAndWriteTemplate(filename, templateName, templateBody, templateData, format.Source)
}

func (g *Generator) Errorf(format string, a ...interface{}) {
	g.ui.Error(fmt.Sprintf(format, a...))
}

func (g *Generator) Fatalf(format string, a ...interface{}) {
	g.Errorf(format, a...)
	os.Exit(1)
}

func (g *Generator) Infof(format string, a ...interface{}) {
	g.ui.Info(fmt.Sprintf(format, a...))
}
