//go:build generate
// +build generate

package main

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"fmt"
	"go/format"
	"log"
	"os"
	"sort"
	"text/template"

	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	filename      = `config_gen.go`
	namesDataFile = "../../names/names_data.csv"
)

type ServiceDatum struct {
	SDKVersion        string
	GoPackage         string
	ProviderNameUpper string
}

type TemplateData struct {
	Services []ServiceDatum
}

func main() {
	fmt.Printf("Generating internal/conns/%s\n", filename)

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

		if l[names.ColExclude] != "" || l[names.ColSkipClientGenerate] != "" {
			continue
		}

		if l[names.ColProviderPackageActual] == "" && l[names.ColProviderPackageCorrect] == "" {
			continue
		}

		if l[names.ColClientSDKV1] != "" {
			td.Services = append(td.Services, ServiceDatum{
				ProviderNameUpper: l[names.ColProviderNameUpper],
				SDKVersion:        "1",
				GoPackage:         l[names.ColGoV1Package],
			})
		}
		// if l[names.ColClientSDKV2] != "" {
		// 	sd := ServiceDatum{
		// 		ProviderNameUpper: l[names.ColProviderNameUpper],
		// 		SDKVersion:        "2",
		// 		GoPackage:         l[names.ColGoV2Package],
		// 	}
		// 	if l[names.ColClientSDKV1] != "" {
		// 		// Use `sdkv2` instead of `v2` to prevent collisions with e.g., `elbv2`
		// 		sd.GoPackage = fmt.Sprintf("%s_sdkv2", l[names.ColGoV2Package])
		// 	}
		// 	td.Services = append(td.Services, sd)
		// }
	}

	sort.SliceStable(td.Services, func(i, j int) bool {
		return td.Services[i].ProviderNameUpper < td.Services[j].ProviderNameUpper
	})

	writeTemplate(tmpl, "awsclient", td)
}

func writeTemplate(body string, templateName string, td TemplateData) {
	f, err := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
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

	contents, err := format.Source(buffer.Bytes())
	if err != nil {
		log.Fatalf("error formatting generated file: %s", err)
	}

	if _, err := f.Write(contents); err != nil {
		f.Close()
		log.Fatalf("error writing to file (%s): %s", filename, err)
	}

	if err := f.Close(); err != nil {
		log.Fatalf("error closing file (%s): %s", filename, err)
	}
}

//go:embed file.tmpl
var tmpl string
