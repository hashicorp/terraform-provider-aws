// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package list

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hashicorp/terraform-provider-aws/names"
	"github.com/hashicorp/terraform-provider-aws/names/data"
	"github.com/hashicorp/terraform-provider-aws/skaff/convert"
)

//go:embed list_common.gtpl
var listCommonTmpl string

//go:embed list_framework.gtpl
var listTmplFramework string

//go:embed list_sdkv2.gtpl
var listTmplSdkV2 string

//go:embed listtest.gtpl
var listTestTmpl string

//go:embed testconfig.gtpl
var lisTestConfigTmpl string

//go:embed query.gtpl
var queryTmpl string

//go:embed websitedoc.gtpl
var websiteTmpl string

type TemplateData struct {
	ListResource              string
	ListResourceAWS           string
	ListResourceLower         string
	ListResourceLowerCamel    string
	ListResourceSnake         string
	IncludeComments           bool
	HumanFriendlyService      string
	HumanFriendlyServiceShort string
	SDKPackage                string
	ServicePackage            string
	Service                   string
	ServiceLower              string
	HumanListResourceName     string
	ProviderResourceName      string
	ARNNamespace              string
}

func Create(listName, snakeName string, comments, framework, force bool) error {
	wd, err := os.Getwd() // os.Getenv("GOPACKAGE") not available since this is not run with go generate
	if err != nil {
		return fmt.Errorf("error reading working directory: %s", err)
	}

	servicePackage := filepath.Base(wd)

	if listName == "" {
		return fmt.Errorf("error checking: no name given")
	}

	if listName == strings.ToLower(listName) {
		return fmt.Errorf("error checking: name should be properly capitalized (e.g., DBInstance)")
	}

	if snakeName != "" && snakeName != strings.ToLower(snakeName) {
		return fmt.Errorf("error checking: snake name should be all lower case with underscores, if needed (e.g., db_instance)")
	}

	if snakeName == "" {
		snakeName = names.ToSnakeCase(listName)
	}

	service, err := data.LookupService(servicePackage)
	if err != nil {
		return fmt.Errorf("error looking up service package data for %q: %w", servicePackage, err)
	}

	templateData := TemplateData{
		ListResource:              listName,
		ListResourceAWS:           convert.ToAWSCapitalization(listName),
		ListResourceLower:         strings.ToLower(listName),
		ListResourceLowerCamel:    convert.ToLowercasePrefix(listName),
		ListResourceSnake:         snakeName,
		HumanFriendlyService:      service.HumanFriendly(),
		HumanFriendlyServiceShort: service.HumanFriendlyShort(),
		IncludeComments:           comments,
		SDKPackage:                service.GoV2Package(),
		ServicePackage:            servicePackage,
		Service:                   service.ProviderNameUpper(),
		ServiceLower:              strings.ToLower(service.ProviderNameUpper()),
		HumanListResourceName:     convert.ToHumanResName(listName),
		ProviderResourceName:      convert.ToProviderResourceName(servicePackage, snakeName),
		ARNNamespace:              service.ARNNamespace(),
	}

	if err := writeListResource(snakeName, framework, force, templateData); err != nil {
		return fmt.Errorf("creating list resource: %w", err)
	}

	tf := fmt.Sprintf("%s_list_test.go", snakeName)
	if err = writeTemplate("listtest", tf, listTestTmpl, force, templateData); err != nil {
		return fmt.Errorf("writing list resource test template: %w", err)
	}

	if err := testConfig(listName, "list_basic", force, templateData, false, false); err != nil {
		return err
	}

	if err := testConfig(listName, "list_include_resource", force, templateData, true, false); err != nil {
		return err
	}

	if err := testConfig(listName, "list_region_override", force, templateData, false, true); err != nil {
		return err
	}

	wf := fmt.Sprintf("%s_%s.html.markdown", servicePackage, snakeName)
	wf = filepath.Join("..", "..", "..", "website", "docs", "list-resources", wf)
	if err = writeTemplate("webdoc", wf, websiteTmpl, force, templateData); err != nil {
		return fmt.Errorf("writing list resource website doc template: %w", err)
	}

	return nil
}

type testConfigTemplateData struct {
	IsIncludeResource bool
	IsRegionOverride  bool
	TemplateData
}

func testConfig(listName, path string, force bool, templateData TemplateData, includeResource, regionOverride bool) error {
	tcf := "main.tf"
	tcf = filepath.Join("testdata", listName, path, tcf)
	if err := os.MkdirAll(filepath.Dir(tcf), 0755); err != nil {
		return fmt.Errorf("creating test config directory: %w", err)
	}

	testConfig := testConfigTemplateData{
		IsIncludeResource: includeResource,
		IsRegionOverride:  regionOverride,
		TemplateData:      templateData,
	}

	if err := writeTemplate("testconfig", tcf, lisTestConfigTmpl, force, testConfig); err != nil {
		return fmt.Errorf("writing list resource test config template: %w", err)
	}

	qf := "main.tfquery.hcl"
	qf = filepath.Join("testdata", listName, path, "query.tfquery.hcl")
	if err := writeTemplate("queryconfig", qf, queryTmpl, force, testConfig); err != nil {
		return fmt.Errorf("writing list resource query config template: %w", err)
	}

	return nil
}

func writeTemplate(templateName, filename, tmpl string, force bool, td any) error {
	tplate, err := template.New(templateName).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("error parsing template: %s", err)
	}

	var buffer bytes.Buffer
	err = tplate.Execute(&buffer, td)
	if err != nil {
		return fmt.Errorf("error executing template: %s", err)
	}

	return write(filename, force, &buffer)
}

func executeTemplate(tmpl *template.Template, templateData any) (io.Reader, error) {
	var buffer bytes.Buffer
	err := tmpl.Execute(&buffer, templateData)

	if err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	return &buffer, nil
}

func write(filename string, force bool, reader io.Reader) error {
	if _, err := os.Stat(filename); !errors.Is(err, fs.ErrNotExist) && !force {
		return fmt.Errorf("file (%s) already exists and force is not set", filename)
	}

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening file (%s): %s", filename, err)
	}

	if _, err := f.ReadFrom(reader); err != nil {
		f.Close() // ignore error; Write error takes precedence
		return fmt.Errorf("error writing to file (%s): %s", filename, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("error closing file (%s): %s", filename, err)
	}

	return nil
}

func writeListResource(snakeName string, framework, force bool, templateData TemplateData) error {
	f := fmt.Sprintf("%s_list.go", snakeName)
	templates, err := template.New("newlist").Parse(listCommonTmpl)
	if err != nil {
		return fmt.Errorf("parsing common template %q: %w", "list_common.gtpl", err)
	}

	tmpl := listTmplFramework
	if !framework {
		tmpl = listTmplSdkV2
	}
	templates, err = templates.Parse(tmpl)
	if err != nil {
		tmplName := "list_framework.gtpl"
		if !framework {
			tmplName = "list_sdkv2.gtpl"
		}
		return fmt.Errorf("parsing implementation template %q: %w", tmplName, err)
	}

	b, err := executeTemplate(templates, templateData)
	if err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	if err = write(f, force, b); err != nil {
		return fmt.Errorf("writing file %q: %w", f, err)
	}

	return nil
}
