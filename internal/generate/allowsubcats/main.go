//go:build generate
// +build generate

package main

import (
	"bytes"
	"encoding/csv"
	"log"
	"os"
	"sort"
	"text/template"
)

const filename = `../../../website/allowed-subcategories.txt`

type ServiceDatum struct {
	HumanFriendly string
}

type TemplateData struct {
	Services []ServiceDatum
}

const (
	// column indices of CSV
	//awsCLIV2Command         = 0
	//awsCLIV2CommandNoDashes = 1
	//goV1Package             = 2
	//goV2Package             = 3
	//providerPackageActual   = 4
	//providerPackageCorrect  = 5
	//aliases                 = 6
	//providerNameUpper       = 7
	//goV1ClientName          = 8
	//skipClientGenerate      = 9
	//sdkVersion              = 10
	//resourcePrefixActual    = 11
	//resourcePrefixCorrect   = 12
	//substituteDocsPrefixes  = 13
	//humanFriendly           = 14
	//brand                   = 15
	//exclude                 = 16
	//allowedSubcategory      = 17
	//deprecatedEnvVar        = 18
	//envVar                  = 19
	//note                    = 20
	humanFriendly      = 14
	exclude            = 16
	allowedSubcategory = 17
)

func main() {
	f, err := os.Open("../../../names/names_data.csv")
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

		if l[allowedSubcategory] == "" || l[exclude] != "" {
			continue
		}

		sd := ServiceDatum{
			HumanFriendly: l[humanFriendly],
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
