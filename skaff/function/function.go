// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-provider-aws/skaff/convert"
)

//go:embed function.tmpl
var functionTmpl string

//go:embed functiontest.tmpl
var functionTestTmpl string

//go:embed websitedoc.tmpl
var websiteTmpl string

var snakeCaseRegex = regexache.MustCompile(`[a-z0-9_]*`)

type TemplateData struct {
	Function        string
	FunctionLower   string
	FunctionSnake   string
	Description     string
	IncludeComments bool
}

func Create(name, snakeName, description string, comments, force bool) error {
	if name == strings.ToLower(name) {
		return fmt.Errorf("name should be properly capitalized (e.g., ARNBuild)")
	}

	if snakeName != "" && snakeName != strings.ToLower(snakeName) {
		return fmt.Errorf("snake name should be all lower case with underscores, if needed (e.g., arn_build)")
	}

	snakeName = convert.ToSnakeCase(name, snakeName)

	templateData := TemplateData{
		Function:        name,
		FunctionLower:   convert.ToLowercasePrefix(name),
		FunctionSnake:   snakeName,
		Description:     description,
		IncludeComments: comments,
	}

	tmpl := functionTmpl
	f := fmt.Sprintf("%s.go", snakeName)
	if err := writeTemplate("function", f, tmpl, force, templateData); err != nil {
		return fmt.Errorf("writing resource template: %w", err)
	}

	tf := fmt.Sprintf("%s_test.go", snakeName)
	if err := writeTemplate("functiontest", tf, functionTestTmpl, force, templateData); err != nil {
		return fmt.Errorf("writing resource test template: %w", err)
	}

	wf := fmt.Sprintf("%s.html.markdown", snakeName)
	wf = filepath.Join("..", "..", "website", "docs", "functions", wf)
	if err := writeTemplate("webdoc", wf, websiteTmpl, force, templateData); err != nil {
		return fmt.Errorf("writing resource website doc template: %w", err)
	}

	return nil
}

func writeTemplate(templateName, filename, tmpl string, force bool, td TemplateData) error {
	if _, err := os.Stat(filename); !errors.Is(err, fs.ErrNotExist) && !force {
		return fmt.Errorf("file (%s) already exists and force is not set", filename)
	}

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening file (%s): %s", filename, err)
	}

	tplate, err := template.New(templateName).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("error parsing template: %s", err)
	}

	var buffer bytes.Buffer
	err = tplate.Execute(&buffer, td)
	if err != nil {
		return fmt.Errorf("error executing template: %s", err)
	}

	if _, err := f.Write(buffer.Bytes()); err != nil {
		f.Close() // ignore error; Write error takes precedence
		return fmt.Errorf("error writing to file (%s): %s", filename, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("error closing file (%s): %s", filename, err)
	}

	return nil
}
