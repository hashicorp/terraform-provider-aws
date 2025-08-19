// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ephemeral

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
	"github.com/hashicorp/terraform-provider-aws/names/data"
	"github.com/hashicorp/terraform-provider-aws/skaff/convert"
)

//go:embed ephemeral.gtpl
var ephemeralTmpl string

//go:embed ephemeraltest.gtpl
var ephemeralTestTmpl string

//go:embed websitedoc.gtpl
var websiteTmpl string

type TemplateData struct {
	EphemeralResource          string
	EphemeralResourceLower     string
	EphemeralResourceSnake     string
	IncludeComments            bool
	HumanFriendlyService       string
	SDKPackage                 string
	ServicePackage             string
	Service                    string
	ServiceLower               string
	AWSServiceName             string
	HumanEphemeralResourceName string
	ProviderResourceName       string
}

func Create(ephemeralName, snakeName string, comments, force bool) error {
	wd, err := os.Getwd() // os.Getenv("GOPACKAGE") not available since this is not run with go generate
	if err != nil {
		return fmt.Errorf("error reading working directory: %s", err)
	}

	servicePackage := filepath.Base(wd)

	if ephemeralName == "" {
		return fmt.Errorf("error checking: no name given")
	}

	if ephemeralName == strings.ToLower(ephemeralName) {
		return fmt.Errorf("error checking: name should be properly capitalized (e.g., DBInstance)")
	}

	if snakeName != "" && snakeName != strings.ToLower(snakeName) {
		return fmt.Errorf("error checking: snake name should be all lower case with underscores, if needed (e.g., db_instance)")
	}

	if snakeName == "" {
		snakeName = names.ToSnakeCase(ephemeralName)
	}

	service, err := data.LookupService(servicePackage)
	if err != nil {
		return fmt.Errorf("error looking up service package data for %q: %w", servicePackage, err)
	}

	templateData := TemplateData{
		EphemeralResource:          ephemeralName,
		EphemeralResourceLower:     strings.ToLower(ephemeralName),
		EphemeralResourceSnake:     snakeName,
		HumanFriendlyService:       service.HumanFriendly(),
		IncludeComments:            comments,
		SDKPackage:                 service.GoV2Package(),
		ServicePackage:             servicePackage,
		Service:                    service.ProviderNameUpper(),
		ServiceLower:               strings.ToLower(service.ProviderNameUpper()),
		AWSServiceName:             service.FullHumanFriendly(),
		HumanEphemeralResourceName: convert.ToHumanResName(ephemeralName),
		ProviderResourceName:       convert.ToProviderResourceName(servicePackage, snakeName),
	}

	tmpl := ephemeralTmpl
	f := fmt.Sprintf("%s_ephemeral.go", snakeName)
	if err = writeTemplate("newephemeral", f, tmpl, force, templateData); err != nil {
		return fmt.Errorf("writing ephemeral resource template: %w", err)
	}

	tf := fmt.Sprintf("%s_ephemeral_test.go", snakeName)
	if err = writeTemplate("ephemeraltest", tf, ephemeralTestTmpl, force, templateData); err != nil {
		return fmt.Errorf("writing ephemeral resource test template: %w", err)
	}

	wf := fmt.Sprintf("%s_%s.html.markdown", servicePackage, snakeName)
	wf = filepath.Join("..", "..", "..", "website", "docs", "ephemeral-resources", wf)
	if err = writeTemplate("webdoc", wf, websiteTmpl, force, templateData); err != nil {
		return fmt.Errorf("writing ephemeral resource website doc template: %w", err)
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
