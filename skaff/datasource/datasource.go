package datasource

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/hashicorp/terraform-provider-aws/names"
)

//go:embed datasource.tmpl
var datasourceTmpl string

//go:embed datasourcetest.tmpl
var datasourceTestTmpl string

//go:embed websitedoc.tmpl
var websiteTmpl string

type TemplateData struct {
	DataSource      string
	DataSourceLower string
	IncludeComments bool
	ServicePackage  string
	Service         string
	ServiceLower    string
	AWSServiceName  string
}

func toSnakeCase(upper string, snakeName string) string {
	if snakeName != "" {
		return snakeName
	}

	re := regexp.MustCompile(`([a-z])([A-Z]{2,})`)
	upper = re.ReplaceAllString(upper, `${1}_${2}`)

	re2 := regexp.MustCompile(`([A-Z][a-z])`)
	return strings.TrimPrefix(strings.ToLower(re2.ReplaceAllString(upper, `_$1`)), "_")
}

func Create(dsName, snakeName string, comments, force bool) error {
	wd, err := os.Getwd() // os.Getenv("GOPACKAGE") not available since this is not run with go generate
	if err != nil {
		return fmt.Errorf("error reading working directory: %s", err)
	}

	servicePackage := filepath.Base(wd)

	if dsName == "" {
		return fmt.Errorf("error checking: no name given")
	}

	if dsName == strings.ToLower(dsName) {
		return fmt.Errorf("error checking: name should be properly capitalized (e.g., DBInstance)")
	}

	if snakeName != "" && snakeName != strings.ToLower(snakeName) {
		return fmt.Errorf("error checking: snake name should be all lower case with underscores, if needed (e.g., db_instance)")
	}

	s, err := names.ProviderNameUpper(servicePackage)
	if err != nil {
		return fmt.Errorf("error getting service connection name: %w", err)
	}

	sn, err := names.FullHumanFriendly(servicePackage)
	if err != nil {
		return fmt.Errorf("error getting AWS service name: %w", err)
	}

	templateData := TemplateData{
		DataSource:      dsName,
		DataSourceLower: strings.ToLower(dsName),
		IncludeComments: comments,
		ServicePackage:  servicePackage,
		Service:         s,
		ServiceLower:    strings.ToLower(s),
		AWSServiceName:  sn,
	}

	f := fmt.Sprintf("%s_data_source.go", toSnakeCase(dsName, snakeName))
	if err = writeTemplate("newds", f, datasourceTmpl, force, templateData); err != nil {
		return fmt.Errorf("writing datasource template: %w", err)
	}

	tf := fmt.Sprintf("%s_data_source_test.go", toSnakeCase(dsName, snakeName))
	if err = writeTemplate("dstest", tf, datasourceTestTmpl, force, templateData); err != nil {
		return fmt.Errorf("writing datasource test template: %w", err)
	}

	wf := fmt.Sprintf("%s_%s.html.markdown", servicePackage, toSnakeCase(dsName, snakeName))
	wf = filepath.Join("..", "..", "..", "website", "docs", "d", wf)
	if err = writeTemplate("webdoc", wf, websiteTmpl, force, templateData); err != nil {
		return fmt.Errorf("writing datasource website doc template: %w", err)
	}

	return nil
}

func writeTemplate(templateName, filename, tmpl string, force bool, td TemplateData) error {
	if _, err := os.Stat(filename); !os.IsNotExist(err) && !force {
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
