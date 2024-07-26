// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resource

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

	"github.com/hashicorp/terraform-provider-aws/names"
	"github.com/hashicorp/terraform-provider-aws/skaff/convert"
)

//go:embed resource.tmpl
var resourceTmpl string

//go:embed resourcefw.tmpl
var resourceFrameworkTmpl string

//go:embed resourcetest.tmpl
var resourceTestTmpl string

//go:embed websitedoc.tmpl
var websiteTmpl string

type TemplateData struct {
	Resource             string
	ResourceLower        string
	ResourceSnake        string
	HumanFriendlyService string
	IncludeComments      bool
	IncludeTags          bool
	ServicePackage       string
	Service              string
	ServiceLower         string
	AWSServiceName       string
	AWSGoSDKV2           bool
	PluginFramework      bool
	HumanResourceName    string
	ProviderResourceName string
}

func Create(resName, snakeName string, comments, force, v2, pluginFramework, tags bool) error {
	wd, err := os.Getwd() // os.Getenv("GOPACKAGE") not available since this is not run with go generate
	if err != nil {
		return fmt.Errorf("error reading working directory: %s", err)
	}

	servicePackage := filepath.Base(wd)

	if resName == "" {
		return fmt.Errorf("error checking: no name given")
	}

	if resName == strings.ToLower(resName) {
		return fmt.Errorf("error checking: name should be properly capitalized (e.g., DBInstance)")
	}

	if snakeName != "" && snakeName != strings.ToLower(snakeName) {
		return fmt.Errorf("error checking: snake name should be all lower case with underscores, if needed (e.g., db_instance)")
	}

	snakeName = convert.ToSnakeCase(resName, snakeName)

	s, err := names.ProviderNameUpper(servicePackage)
	if err != nil {
		return fmt.Errorf("error getting service connection name: %w", err)
	}

	sn, err := names.FullHumanFriendly(servicePackage)
	if err != nil {
		return fmt.Errorf("error getting AWS service name: %w", err)
	}

	hf, err := names.HumanFriendly(servicePackage)
	if err != nil {
		return fmt.Errorf("error getting human-friendly name: %w", err)
	}

	templateData := TemplateData{
		Resource:             resName,
		ResourceLower:        strings.ToLower(resName),
		ResourceSnake:        snakeName,
		HumanFriendlyService: hf,
		IncludeComments:      comments,
		IncludeTags:          tags,
		ServicePackage:       servicePackage,
		Service:              s,
		ServiceLower:         strings.ToLower(s),
		AWSServiceName:       sn,
		AWSGoSDKV2:           v2,
		PluginFramework:      pluginFramework,
		HumanResourceName:    convert.ToHumanResName(resName),
		ProviderResourceName: convert.ToProviderResourceName(servicePackage, snakeName),
	}

	tmpl := resourceTmpl
	if pluginFramework {
		tmpl = resourceFrameworkTmpl
	}
	f := fmt.Sprintf("%s.go", snakeName)
	if err = writeTemplate("newres", f, tmpl, force, templateData); err != nil {
		return fmt.Errorf("writing resource template: %w", err)
	}

	tf := fmt.Sprintf("%s_test.go", snakeName)
	if err = writeTemplate("restest", tf, resourceTestTmpl, force, templateData); err != nil {
		return fmt.Errorf("writing resource test template: %w", err)
	}

	wf := fmt.Sprintf("%s_%s.html.markdown", servicePackage, snakeName)
	wf = filepath.Join("..", "..", "..", "website", "docs", "r", wf)
	if err = writeTemplate("webdoc", wf, websiteTmpl, force, templateData); err != nil {
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

	//contents, err := format.Source(buffer.Bytes())
	//if err != nil {
	//	return fmt.Errorf("error formatting generated file: %s", err)
	//}

	//if _, err := f.Write(contents); err != nil {
	if _, err := f.Write(buffer.Bytes()); err != nil {
		f.Close() // ignore error; Write error takes precedence
		return fmt.Errorf("error writing to file (%s): %s", filename, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("error closing file (%s): %s", filename, err)
	}

	return nil
}
