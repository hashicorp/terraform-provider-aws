//go:build generate
// +build generate

package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	filename      = `../../../website/allowed-subcategories.txt`
	namesDataFile = "../../../names/names_data.csv"
)

type ServiceDatum struct {
	HumanFriendly string
}

type TemplateData struct {
	Services []ServiceDatum
}

func main() {
	fmt.Printf("Generating %s\n", strings.TrimPrefix(filename, "../../../"))

	f, err := os.Open(namesDataFile)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	csvReader := csv.NewReader(f)

	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	td := TemplateData{}

	for i, l := range data {
		if i < 1 { // no header
			continue
		}

		if l[names.ColExclude] != "" && l[names.ColAllowedSubcategory] == "" {
			continue
		}

		sd := ServiceDatum{
			HumanFriendly: l[names.ColHumanFriendly],
		}

		td.Services = append(td.Services, sd)
	}

	sort.SliceStable(td.Services, func(i, j int) bool {
		return td.Services[i].HumanFriendly < td.Services[j].HumanFriendly
	})

	writeTemplate(tmpl, "allowsubcats", td)
}

func writeTemplate(body string, templateName string, td TemplateData) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening file (%s): %s", filename, err)
	}

	tplate, err := template.New(templateName).Parse(body)
	if err != nil {
		log.Fatalf("error parsing template: %s", err)
	}

	var buffer bytes.Buffer
	err = tplate.Execute(&buffer, td)
	if err != nil {
		log.Fatalf("error executing template: %s", err)
	}

	if _, err := f.Write(buffer.Bytes()); err != nil {
		f.Close()
		log.Fatalf("error writing to file (%s): %s", filename, err)
	}

	if err := f.Close(); err != nil {
		log.Fatalf("error closing file (%s): %s", filename, err)
	}
}

var tmpl = `
{{- range .Services }}
{{ .HumanFriendly }}
{{- end }}
`
