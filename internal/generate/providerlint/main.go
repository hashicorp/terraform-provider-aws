//go:build generate
// +build generate

package main

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/hashicorp/terraform-provider-aws/names"
)

//go:embed header.tmpl
var header string

const (
	filename      = `../../../.github/workflows/providerlint.yml`
	namesDataFile = "../../../names/names_data.csv"
	jobs          = 5
)

type TemplateData struct {
	ServiceChunks []string
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

	var services []string

	for i, l := range data {
		if i < 1 { // no header
			continue
		}

		if l[names.ColExclude] != "" && l[names.ColAllowedSubcategory] == "" {
			continue
		}

		if l[names.ColProviderPackageActual] == "" && l[names.ColProviderPackageCorrect] == "" {
			continue
		}

		p := l[names.ColProviderPackageCorrect]

		if l[names.ColProviderPackageActual] != "" {
			p = l[names.ColProviderPackageActual]
		}

		rp := p

		if l[names.ColSplitPackageRealPackage] != "" {
			rp = l[names.ColSplitPackageRealPackage]
		}

		if _, err := os.Stat(fmt.Sprintf("../../service/%s", rp)); err != nil || os.IsNotExist(err) {
			continue
		}

		services = append(services, rp)
	}

	sort.Strings(services)

	perJob := int(math.Ceil(float64(len(services)) / jobs))

	td := TemplateData{}

	for i := 0; i < jobs; i++ {
		if i == (jobs - 1) {
			td.ServiceChunks = append(td.ServiceChunks, strings.Join(services[i*perJob:], " "))
			continue
		}

		td.ServiceChunks = append(td.ServiceChunks, strings.Join(services[i*perJob:((i+1)*perJob)], " "))
	}

	writeTemplate(fmt.Sprintf("%s%s", header, tmpl), "providerlint", td)
}

func writeTemplate(body string, templateName string, td TemplateData) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("error opening file (%s): %s", filename, err)
	}

	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
		"ord": func(i int) string {
			return string(97 + i)
		},
	}

	tplate, err := template.New(templateName).Funcs(funcMap).Parse(body)
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

var tmpl = `{{ $length := len .ServiceChunks }}{{- range $i, $chunk := .ServiceChunks }}
  providerlint{{ ord $i }}:
    name: {{ inc $i }} of {{ $length }}
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version-file: .go-version
      - name: go env
        run: echo "GOCACHE=$(go env GOCACHE)" >> $GITHUB_ENV
      - uses: actions/cache@v3
        continue-on-error: true
        timeout-minutes: 2
        with:
          path: ${{"{{"}} env.GOCACHE {{"}}"}}
          key: ${{"{{"}} runner.os {{"}}"}}-GOCACHE-${{"{{"}} hashFiles('go.sum') {{"}}"}}-${{"{{"}} hashFiles('internal/**') {{"}}"}}
      - uses: actions/cache@v3
        continue-on-error: true
        timeout-minutes: 2
        with:
          path: ~/go/pkg/mod
          key: ${{"{{"}} runner.os {{"}}"}}-go-pkg-mod-${{"{{"}} hashFiles('go.sum') {{"}}"}}
      - run: cd .ci/providerlint && go install .
      - run: .ci/scripts/providerlint.sh {{ $chunk }}
{{- end }}
`
