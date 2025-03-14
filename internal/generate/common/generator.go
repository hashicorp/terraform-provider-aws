// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"bytes"
	"fmt"
	"go/format"
	"maps"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	"github.com/hashicorp/cli"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

func (g *Generator) UI() cli.Ui {
	return g.ui
}

func (g *Generator) Infof(format string, a ...any) {
	g.ui.Info(fmt.Sprintf(format, a...))
}

func (g *Generator) Warnf(format string, a ...any) {
	g.ui.Warn(fmt.Sprintf(format, a...))
}

func (g *Generator) Errorf(format string, a ...any) {
	g.ui.Error(fmt.Sprintf(format, a...))
}

func (g *Generator) Fatalf(format string, a ...any) {
	g.Errorf(format, a...)
	os.Exit(1)
}

type Destination interface {
	CreateDirectories() error
	Write() error
	BufferBytes(body []byte) error
	BufferTemplate(templateName, templateBody string, templateData any, funcMaps ...template.FuncMap) error
	BufferTemplateSet(templates *template.Template, templateData any) error
}

// NewGoFileDestination creates a new destination for a Go file with the given name and with Go code
// formatting. The file will be created if it does not exist, and truncated if it does. The formatting
// is done with gofmt and goimports to fix many common formatting issues and adding and removing imports.
// This provides a degree of freedom in templates where it can be difficult to determine the correct
// imports. This allows you to simplify templates since you can over include imports in case they are
// needed, knowing that goimports will remove any unnecessary packages.
func (g *Generator) NewGoFileDestination(filename string) Destination {
	return &fileDestination{
		baseDestination: baseDestination{
			formatter:      format.Source,
			writeFormatter: goodgo,
		},
		filename: filename,
	}
}

func (g *Generator) NewUnformattedFileDestination(filename string) Destination {
	return &fileDestination{
		filename: filename,
	}
}

type fileDestination struct {
	baseDestination
	append   bool
	filename string
}

func (d *fileDestination) CreateDirectories() error {
	const (
		perm os.FileMode = 0755
	)
	dirname := path.Dir(d.filename)
	err := os.MkdirAll(dirname, perm)

	if err != nil {
		return fmt.Errorf("creating target directory %s: %w", dirname, err)
	}

	return nil
}

// Write writes the buffer to an actual disk file, as opposed to writing to memory like BufferBytes or
// BufferTemplate.
func (d *fileDestination) Write() error {
	var flags int
	if d.append {
		flags = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	} else {
		flags = os.O_TRUNC | os.O_CREATE | os.O_WRONLY
	}

	f, err := os.OpenFile(d.filename, flags, 0644) //nolint:mnd // good protection for new files
	if err != nil {
		return fmt.Errorf("opening file (%s): %w", d.filename, err)
	}
	defer f.Close()

	content := d.buffer.String()
	if d.writeFormatter != nil {
		formattedContent, err := d.writeFormatter([]byte(content))
		if err != nil {
			return fmt.Errorf("formatting written template:\n%s\n%w", content, err)
		}
		content = string(formattedContent)
	}

	_, err = f.WriteString(content)
	if err != nil {
		return fmt.Errorf("writing to file (%s): %w", d.filename, err)
	}

	return nil
}

type baseDestination struct {
	formatter      func([]byte) ([]byte, error)
	writeFormatter func([]byte) ([]byte, error)
	buffer         strings.Builder
}

// BufferBytes buffers the given raw bytes.
func (d *baseDestination) BufferBytes(body []byte) error {
	_, err := d.buffer.Write(body)
	return err
}

// BufferTemplate parses and executes the template with the given data, applying any
// formatter previously set up, such as Go code formatting, and buffers the result.
func (d *baseDestination) BufferTemplate(templateName, templateBody string, templateData any, funcMaps ...template.FuncMap) error {
	body, err := parseTemplate(templateName, templateBody, templateData, funcMaps...)

	if err != nil {
		return err
	}

	body, err = d.format(body)
	if err != nil {
		return err
	}

	return d.BufferBytes(body)
}

func parseTemplate(templateName, templateBody string, templateData any, funcMaps ...template.FuncMap) ([]byte, error) {
	funcMap := template.FuncMap{
		// FirstUpper returns a string with the first character as upper case.
		"FirstUpper": func(s string) string {
			if s == "" {
				return ""
			}
			r, n := utf8.DecodeRuneInString(s)
			return string(unicode.ToUpper(r)) + s[n:]
		},
		// Title returns a string with the first character of each word as upper case.
		"Title": cases.Title(language.Und, cases.NoLower).String,
	}
	for _, v := range funcMaps {
		maps.Copy(funcMap, v) // Extras overwrite defaults.
	}
	tmpl, err := template.New(templateName).Funcs(funcMap).Parse(templateBody)

	if err != nil {
		return nil, fmt.Errorf("parsing function template: %w", err)
	}

	return executeTemplate(tmpl, templateData)
}

func executeTemplate(tmpl *template.Template, templateData any) ([]byte, error) {
	var buffer bytes.Buffer
	err := tmpl.Execute(&buffer, templateData)

	if err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	return buffer.Bytes(), nil
}

// BufferTemplateSet executes the templates with the given data, applying any
// formatter previously set up, such as Go code formatting, and buffers the result.
func (d *baseDestination) BufferTemplateSet(templates *template.Template, templateData any) error {
	body, err := executeTemplate(templates, templateData)
	if err != nil {
		return err
	}

	body, err = d.format(body)
	if err != nil {
		return err
	}

	return d.BufferBytes(body)
}

func (d *baseDestination) format(body []byte) ([]byte, error) {
	if d.formatter == nil {
		return body, nil
	}

	unformattedBody := body
	body, err := d.formatter(unformattedBody)
	if err != nil {
		return nil, fmt.Errorf("formatting parsed template:\n%s\n%w", unformattedBody, err)
	}

	return body, nil
}

// goodgo formats the given Go source code using gofmt and goimports.
func goodgo(body []byte) ([]byte, error) {
	// Run gofmt with the -s option
	formattedBody, err := runCommand("gofmt", "-s", body)
	if err != nil {
		return nil, fmt.Errorf("running gofmt: %w", err)
	}

	// Run goimports to fix imports
	formattedBody, err = runCommand("goimports", "-v", formattedBody)
	if err != nil {
		return nil, fmt.Errorf("running goimports: %w", err)
	}

	return formattedBody, nil
}

// runCommand runs a command with the given arguments and input, and returns the output.
func runCommand(name string, arg string, input []byte) ([]byte, error) {
	cmd := exec.Command(name, arg)
	cmd.Stdin = bytes.NewReader(input)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
