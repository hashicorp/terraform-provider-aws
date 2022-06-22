//go:build generate
// +build generate

package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/hashicorp/terraform-provider-aws/names"
)

//go:embed semgrep_header.tmpl
var header string

//go:embed configs.tmpl
var configs string

const (
	filename            = `../../../.semgrep-service-name.yml`
	filenameCAE         = `../../../.semgrep-caps-aws-ec2.yml`
	filenameConfigs     = `../../../.semgrep-configs.yml`
	namesDataFile       = "../../../names/names_data.csv"
	capsDataFile        = "../../../names/caps.csv"
	maxBadCaps          = 21
	semgrepConfigChunks = 4
)

type ServiceDatum struct {
	ProviderPackage string
	ServiceAlias    string
	LowerAlias      string
	MainAlias       bool
	FilePrefix      string
}

type TemplateData struct {
	Services []ServiceDatum
}

type CAEData struct {
	BadCaps []string
}

func main() {
	fmt.Printf("Generating %s\n", strings.TrimPrefix(filenameCAE, "../../../"))

	badCaps := readBadCaps()

	cd := CAEData{}
	cd.BadCaps = badCaps

	writeCAE(tmplCAE, "caps-aws-ec2", cd)

	fmt.Printf("Generating %s\n", strings.TrimPrefix(filenameConfigs, "../../../"))

	writeConfigs()

	fmt.Printf("Generating %s\n", strings.TrimPrefix(filename, "../../../"))

	td := TemplateData{}

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

		if l[names.ColAliases] != "" {
			for _, v := range strings.Split(l[names.ColAliases], ";") {
				if strings.ToLower(v) == "es" {
					continue // "es" is too short to usefully grep
				}

				if strings.ToLower(v) == "config" {
					continue // "config" is too ubiquitous
				}

				sd := ServiceDatum{
					ProviderPackage: rp,
					ServiceAlias:    v,
					LowerAlias:      strings.ToLower(v),
					MainAlias:       false,
				}

				td.Services = append(td.Services, sd)
			}
		}

		sd := ServiceDatum{
			ProviderPackage: rp,
			ServiceAlias:    l[names.ColProviderNameUpper],
			LowerAlias:      strings.ToLower(p),
			MainAlias:       true,
		}

		if l[names.ColFilePrefix] != "" {
			sd.FilePrefix = l[names.ColFilePrefix]
		}

		td.Services = append(td.Services, sd)
	}

	sort.SliceStable(td.Services, func(i, j int) bool {
		if td.Services[i].LowerAlias == td.Services[j].LowerAlias {
			return len(td.Services[i].ServiceAlias) > len(td.Services[j].ServiceAlias)
		}
		return td.Services[i].LowerAlias < td.Services[j].LowerAlias
	})

	writeTemplate(tmpl, "servicesemgrep", td)

	breakUpBigFile()

	fmt.Printf("  Removing %s\n", strings.TrimPrefix(filename, "../../../"))
	err = os.Remove(filename)
	if err != nil {
		log.Fatal(err)
	}
}

func readBadCaps() []string {
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

	var capsList []string

	for i, row := range caps {
		if i < 1 { // skip header
			continue
		}

		// 0 - wrong
		// 1 - right

		if row[0] == "" {
			continue
		}

		capsList = append(capsList, row[0])
	}

	sort.SliceStable(capsList, func(i, j int) bool {
		if len(capsList[i]) == len(capsList[j]) {
			return capsList[i] < capsList[j]
		}
		return len(capsList[j]) < len(capsList[i])
	})

	var chunks [][]string
	onChunk := -1

	for i, v := range capsList {
		if i%maxBadCaps == 0 {
			onChunk++
			chunks = append(chunks, []string{})
		}

		chunks[onChunk] = append(chunks[onChunk], v)
	}

	var strChunks []string

	for _, v := range chunks {
		strChunks = append(strChunks, strings.Join(v, "|"))
	}

	return strChunks
}

func writeCAE(body string, templateName string, cd CAEData) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(filenameCAE, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("error opening file (%s): %s", filename, err)
	}

	tplate, err := template.New(templateName).Parse(body)
	if err != nil {
		log.Fatalf("error parsing template: %s", err)
	}

	var buffer bytes.Buffer
	err = tplate.Execute(&buffer, cd)
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

func writeConfigs() {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(filenameConfigs, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("error opening file (%s): %s", filename, err)
	}

	if _, err := f.Write([]byte(configs)); err != nil {
		f.Close()
		log.Fatalf("error writing to file (%s): %s", filename, err)
	}

	if err := f.Close(); err != nil {
		log.Fatalf("error closing file (%s): %s", filename, err)
	}
}

func writeTemplate(body string, templateName string, td TemplateData) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
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

func breakUpBigFile() {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	lines, err := lineCounter(f)
	if err != nil {
		log.Fatal(err)
	}

	_, err = f.Seek(0, 0)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(f)

	l := 0
	chunk := 0
	var w *bufio.Writer
	var piece *os.File
	var cfile string
	passedChunk := false

	re := regexp.MustCompile(`^  - id: `)

	for scanner.Scan() {
		if l%(lines/semgrepConfigChunks) == 0 {
			passedChunk = true
		}

		if passedChunk && scanner.Text() != "" && re.MatchString(scanner.Text()) {
			passedChunk = false

			if w != nil {
				w.Flush()
			}

			if piece != nil {
				piece.Close()
			}

			cfile = fmt.Sprintf("%s%d.yml", strings.TrimSuffix(filename, ".yml"), chunk)
			fmt.Printf("  Splitting into %s\n", strings.TrimPrefix(cfile, "../../../"))
			chunk++

			var err error
			piece, err = os.OpenFile(cfile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				log.Fatalf("error opening file (%s): %s", cfile, err)
			}

			w = bufio.NewWriter(piece)
			_, err = w.WriteString(header)
			if err != nil {
				log.Fatalf("error writing header to file (%s): %s", cfile, err)
			}
			w.Flush()
		}

		if w != nil {
			_, err = w.WriteString(fmt.Sprintf("%s\n", scanner.Text()))
			if err != nil {
				log.Fatalf("error writing to file (%s): %s", cfile, err)
			}
		}

		l++
	}

	if w != nil {
		w.Flush()
	}
}

func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

var tmplCAE = `# Generated by internal/generate/servicesemgrep/main.go; DO NOT EDIT.
rules:
  - id: aws-in-func-name
    languages:
      - go
    message: Do not use "AWS" in func name inside AWS Provider
    paths:
      include:
        - internal
    patterns:
      - pattern: func $NAME( ... ) { ... }
      - metavariable-pattern:
          metavariable: $NAME
          patterns:
            - pattern-regex: "(?i)AWS"
    severity: WARNING
  - id: aws-in-const-name
    languages:
      - go
    message: Do not use "AWS" in const name inside AWS Provider
    paths:
      include:
        - internal
    patterns:
      - pattern: const $NAME = ...
      - metavariable-pattern:
          metavariable: $NAME
          patterns:
            - pattern-regex: "(?i)AWS"
    severity: WARNING
  - id: aws-in-var-name
    languages:
      - go
    message: Do not use "AWS" in var name inside AWS Provider
    paths:
      include:
        - internal
    patterns:
      - pattern: var $NAME = ...
      - metavariable-pattern:
          metavariable: $NAME
          patterns:
            - pattern-regex: "(?i)AWS"
    severity: WARNING
{{- range $i, $s := .BadCaps }}
  - id: caps{{- $i }}-in-func-name
    languages:
      - go
    message: Use correct caps in func name (i.e., HTTPS or https, not Https) (see list at https://github.com/hashicorp/terraform-provider-aws/tree/main/names/caps.md)
    paths:
      include:
        - internal
    patterns:
      - pattern: func $NAME( ... ) { ... }
      - metavariable-pattern:
          metavariable: $NAME
          patterns:
            - pattern-regex: "({{ $s }})"
    severity: WARNING
  - id: caps{{- $i }}-in-const-name
    languages:
      - go
    message: Use correct caps in const name (i.e., HTTPS or https, not Https) (see list at https://github.com/hashicorp/terraform-provider-aws/tree/main/names/caps.md)
    paths:
      include:
        - internal
    patterns:
      - pattern: const $NAME = ...
      - metavariable-pattern:
          metavariable: $NAME
          patterns:
            - pattern-regex: "({{ $s }})"
    severity: WARNING
  - id: caps{{- $i }}-in-var-name
    languages:
      - go
    message: Use correct caps in var name (i.e., HTTPS or https, not Https) (see list at https://github.com/hashicorp/terraform-provider-aws/tree/main/names/caps.md)
    paths:
      include:
        - internal
    patterns:
      - pattern: var $NAME = ...
      - metavariable-pattern:
          metavariable: $NAME
          patterns:
          - pattern-regex: "({{ $s }})"
    severity: WARNING
{{- end }}
  - id: ec2-in-func-name
    languages:
      - go
    message: Do not use "EC2" in func name inside ec2 package
    paths:
      include:
        - internal/service/ec2
    patterns:
      - pattern: func $NAME( ... ) { ... }
      - metavariable-pattern:
          metavariable: $NAME
          patterns:
            - pattern-regex: "(?i)EC2"
            - pattern-not-regex: ^TestAcc.*
    severity: WARNING
  - id: ec2-in-const-name
    languages:
      - go
    message: Do not use "EC2" in const name inside ec2 package
    paths:
      include:
        - internal/service/ec2
    patterns:
      - pattern: const $NAME = ...
      - metavariable-pattern:
          metavariable: $NAME
          patterns:
            - pattern-regex: "(?i)EC2"
    severity: WARNING
  - id: ec2-in-var-name
    languages:
      - go
    message: Do not use "EC2" in var name inside ec2 package
    paths:
      include:
        - internal/service/ec2
    patterns:
      - pattern: var $NAME = ...
      - metavariable-pattern:
          metavariable: $NAME
          patterns:
            - pattern-regex: "(?i)EC2"
    severity: WARNING
`

var tmpl = `{{- range .Services }}
{{- if not .FilePrefix }}
  - id: {{ .LowerAlias }}-in-func-name
    languages:
      - go
    message: Do not use "{{ .ServiceAlias }}" in func name inside {{ .ProviderPackage }} package
    paths:
      include:
        - internal/service/{{ .ProviderPackage }}
    patterns:
      - pattern: func $NAME( ... ) { ... }
      - metavariable-pattern:
          metavariable: $NAME
          patterns:
            - pattern-regex: "(?i){{ .ServiceAlias }}"
            {{- if eq .ServiceAlias "Deploy" }}
            - pattern-not-regex: Deployment(Group|Config|Style)
            {{- end }}
            {{- if eq .ServiceAlias "CE" }}
            - pattern-not-regex: ource
            {{- end }}
            {{- if eq .ServiceAlias "Connect" }}
            - pattern-not-regex: .*uickConnect.*
            {{- end }}
            {{- if eq .ServiceAlias "CloudTrail" }}
            - pattern-not-regex: ^testAccCloudTrailConfig_.*
            {{- end }}
            - pattern-not-regex: ^TestAcc.*
    severity: WARNING
{{- end }}
{{- if .MainAlias }}
  - id: {{ .LowerAlias }}-in-test-name
    languages:
      - go
    message: Include "{{ .ServiceAlias }}" in test name
    paths:
      include:
        - internal/service/{{ .ProviderPackage }}/{{ .FilePrefix }}*_test.go
    patterns:
      - pattern: func $NAME( ... ) { ... }
      - metavariable-pattern:
          metavariable: $NAME
          patterns:
            - pattern-not-regex: "^TestAcc{{ .ServiceAlias }}"
            - pattern-regex: ^TestAcc.*
    severity: WARNING
{{- end }}
{{- if not .FilePrefix }}
  - id: {{ .LowerAlias }}-in-const-name
    languages:
      - go
    message: Do not use "{{ .ServiceAlias }}" in const name inside {{ .ProviderPackage }} package
    paths:
      include:
        - internal/service/{{ .ProviderPackage }}
    patterns:
      - pattern: const $NAME = ...
      - metavariable-pattern:
          metavariable: $NAME
          patterns:
            - pattern-regex: "(?i){{ .ServiceAlias }}"
            {{- if eq .ServiceAlias "Connect" }}
            - pattern-not-regex: .*uickConnect.*
            {{- end }}
    severity: WARNING
  - id: {{ .LowerAlias }}-in-var-name
    languages:
      - go
    message: Do not use "{{ .ServiceAlias }}" in var name inside {{ .ProviderPackage }} package
    paths:
      include:
        - internal/service/{{ .ProviderPackage }}
    patterns:
      - pattern: var $NAME = ...
      - metavariable-pattern:
          metavariable: $NAME
          patterns:
            - pattern-regex: "(?i){{ .ServiceAlias }}"
            {{- if eq .ServiceAlias "Connect" }}
            - pattern-not-regex: .*uickConnect.*
            {{- end }}
    severity: WARNING
{{- end }}
{{- end }}
`
