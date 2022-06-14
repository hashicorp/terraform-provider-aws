//go:build generate
// +build generate

package main

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"
)

//go:embed header.tmpl
var header string

const (
	filename     = "../../../names/caps.md"
	capsDataFile = "../../../names/caps.csv"
	maxBadCaps   = 31
)

type CapsDatum struct {
	Wrong string
	Right string
	Test  string
}

type TemplateData struct {
	BadCaps []CapsDatum
}

func main() {
	fmt.Printf("Generating %s\n", strings.TrimPrefix(filename, "../../../"))

	badCaps := readBadCaps()

	td := TemplateData{}
	td.BadCaps = badCaps

	writeTemplate(header+tmpl, "namescapslist", td)
}

func readBadCaps() []CapsDatum {
	cf, err := os.Open(capsDataFile)
	if err != nil {
		log.Fatal(err)
	}

	defer cf.Close()

	csvReader := csv.NewReader(cf)

	caps, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	var capsList []CapsDatum

	for i, row := range caps {
		if i < 1 { // skip header
			continue
		}

		// 0 - wrong
		// 1 - right

		if row[0] == "" {
			continue
		}

		capsList = append(capsList, CapsDatum{
			Wrong: row[0],
			Right: row[1],
		})
	}

	sort.SliceStable(capsList, func(i, j int) bool {
		if len(capsList[i].Wrong) == len(capsList[j].Wrong) {
			return capsList[i].Wrong < capsList[j].Wrong
		}
		return len(capsList[j].Wrong) < len(capsList[i].Wrong)
	})

	onChunk := -1

	for i := range capsList {
		if i%maxBadCaps == 0 {
			onChunk++
		}

		capsList[i].Test = fmt.Sprintf(`%s%d`, "caps", onChunk)
	}

	sort.SliceStable(capsList, func(i, j int) bool {
		if strings.EqualFold(capsList[i].Wrong, capsList[j].Wrong) {
			return capsList[i].Wrong < capsList[j].Wrong
		}
		return strings.ToLower(capsList[i].Wrong) < strings.ToLower(capsList[j].Wrong)
	})

	return capsList
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
The caps enforced are as follows:

| Wrong | Right | Test# |
| --- | --- | --- |
{{- range .BadCaps }}
| {{ .Wrong }} | {{ .Right }} | {{ .Test }} |
{{- end }}
`
