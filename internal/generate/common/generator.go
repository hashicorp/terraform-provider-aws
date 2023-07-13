// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"strings"
	"text/template"

	"github.com/mitchellh/cli"
)

type Generator struct {
	ui cli.Ui
}

type Destination interface {
	Write() error
	WriteBytes(body []byte) error
	WriteTemplate(templateName, templateBody string, templateData any) error
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

func (g *Generator) NewGoFileDestination(filename string) Destination {
	return &fileDestination{
		filename:  filename,
		formatter: format.Source,
	}
}

func (g *Generator) NewUnformattedFileDestination(filename string) Destination {
	return &fileDestination{
		filename:  filename,
		formatter: func(b []byte) ([]byte, error) { return b, nil },
	}
}

type fileDestination struct {
	append    bool
	filename  string
	formatter func([]byte) ([]byte, error)
	buffer    strings.Builder
}

func (d *fileDestination) Write() error {
	var flags int
	if d.append {
		flags = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	} else {
		flags = os.O_TRUNC | os.O_CREATE | os.O_WRONLY
	}
	f, err := os.OpenFile(d.filename, flags, 0644) //nolint:gomnd

	if err != nil {
		return fmt.Errorf("opening file (%s): %w", d.filename, err)
	}

	defer f.Close()

	_, err = f.WriteString(d.buffer.String())

	if err != nil {
		return fmt.Errorf("writing to file (%s): %w", d.filename, err)
	}

	return nil
}

func (d *fileDestination) WriteBytes(body []byte) error {
	_, err := d.buffer.Write(body)
	return err
}

func (d *fileDestination) WriteTemplate(templateName, templateBody string, templateData any) error {
	unformattedBody, err := parseTemplate(templateName, templateBody, templateData)

	if err != nil {
		return err
	}

	body, err := d.formatter(unformattedBody)

	if err != nil {
		return fmt.Errorf("formatting parsed template:\n%s\n%w", unformattedBody, err)
	}

	_, err = d.buffer.Write(body)
	return err
}

func parseTemplate(templateName, templateBody string, templateData any) ([]byte, error) {
	funcMap := template.FuncMap{
		"Title": strings.Title,
	}
	tmpl, err := template.New(templateName).Funcs(funcMap).Parse(templateBody)

	if err != nil {
		return nil, fmt.Errorf("parsing function template: %w", err)
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, templateData)

	if err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	return buffer.Bytes(), nil
}

func (g *Generator) Infof(format string, a ...interface{}) {
	g.ui.Info(fmt.Sprintf(format, a...))
}

func (g *Generator) Warnf(format string, a ...interface{}) {
	g.ui.Warn(fmt.Sprintf(format, a...))
}

func (g *Generator) Errorf(format string, a ...interface{}) {
	g.ui.Error(fmt.Sprintf(format, a...))
}

func (g *Generator) Fatalf(format string, a ...interface{}) {
	g.Errorf(format, a...)
	os.Exit(1)
}
