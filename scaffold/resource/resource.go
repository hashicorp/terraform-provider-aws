package resource

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

//go:embed resource.tmpl
var body string

type TemplateData struct {
	Resource        string
	IncludeComments bool
	ServicePackage  string
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

func Create(name, snakeName string, comments bool) error {
	wd, err := os.Getwd() // os.Getenv("GOPACKAGE") not available since this is not run with go generate
	if err != nil {
		return fmt.Errorf("error reading working directory: %s", err)
	}

	servicePackage := filepath.Base(wd)

	if name == "" {
		return fmt.Errorf("error checking: no name given")
	}

	if name == strings.ToLower(name) {
		return fmt.Errorf("error checking: name should be properly capitalized (e.g., DBInstance)")
	}

	if snakeName != "" && snakeName != strings.ToLower(snakeName) {
		return fmt.Errorf("error checking: snake name should be all lower case with underscores, if needed (e.g., db_instance)")
	}

	filename := fmt.Sprintf("%s.go", toSnakeCase(name, snakeName))

	templateData := TemplateData{
		Resource:        name,
		IncludeComments: comments,
		ServicePackage:  servicePackage,
	}

	return writeTemplate("newres", filename, templateData)
}

func writeTemplate(templateName string, filename string, td TemplateData) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening file (%s): %s", filename, err)
	}

	tplate, err := template.New(templateName).Parse(body)
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
