// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	"bufio"
	"bytes"
	"cmp"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"slices"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names/data"
)

//go:embed semgrep_header.tmpl
var header string

//go:embed configs.tmpl
var configs string

//go:embed cae.tmpl
var tmplCAE string

//go:embed service.tmpl
var tmpl string

const (
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
	const (
		filename        = `../../../.ci/.semgrep-service-name.yml`
		filenameCAE     = `../../../.ci/.semgrep-caps-aws-ec2.yml`
		filenameConfigs = `../../../.ci/.semgrep-configs.yml`
		capsDataFile    = "../../../names/caps.csv"
	)
	g := common.NewGenerator()

	g.Infof("Generating %s", strings.TrimPrefix(filenameCAE, "../../../"))

	badCaps, err := readBadCaps(capsDataFile)

	if err != nil {
		g.Fatalf("error reading %s: %s", capsDataFile, err)
	}

	cd := CAEData{}
	cd.BadCaps = badCaps

	d := g.NewUnformattedFileDestination(filenameCAE)

	if err := d.BufferTemplate("caps-aws-ec2", tmplCAE, cd); err != nil {
		g.Fatalf("generating file (%s): %s", filenameCAE, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filenameCAE, err)
	}

	g.Infof("Generating %s", strings.TrimPrefix(filenameConfigs, "../../../"))

	d = g.NewUnformattedFileDestination(filenameConfigs)

	if err := d.BufferBytes([]byte(configs)); err != nil {
		g.Fatalf("generating file (%s): %s", filenameConfigs, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filenameConfigs, err)
	}

	g.Infof("Generating %s", strings.TrimPrefix(filename, "../../../"))

	data, err := data.ReadAllServiceData()

	if err != nil {
		g.Fatalf("error reading service data: %s", err)
	}

	td := TemplateData{}

	for _, l := range data {
		if l.Exclude() && !l.AllowedSubcategory() {
			continue
		}

		p := l.ProviderPackage()

		rp := p

		if l.SplitPackageRealPackage() != "" {
			rp = l.SplitPackageRealPackage()
		}

		if _, err := os.Stat(fmt.Sprintf("../../service/%s", rp)); err != nil || errors.Is(err, fs.ErrNotExist) {
			continue
		}

		for _, v := range l.Aliases() {
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

		sd := ServiceDatum{
			ProviderPackage: rp,
			ServiceAlias:    l.ProviderNameUpper(),
			LowerAlias:      strings.ToLower(p),
			MainAlias:       true,
		}

		if l.FilePrefix() != "" {
			sd.FilePrefix = l.FilePrefix()
		}

		td.Services = append(td.Services, sd)
	}

	slices.SortStableFunc(td.Services, func(a, b ServiceDatum) int {
		return cmp.Or(
			cmp.Compare(a.LowerAlias, b.LowerAlias),
			// Reverse order
			cmp.Compare(b.ServiceAlias, a.ServiceAlias),
		)
	})

	d = g.NewUnformattedFileDestination(filename)

	if err := d.BufferTemplate("servicesemgrep", tmpl, td); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := breakUpBigFile(g, filename, header); err != nil {
		g.Fatalf("error: %s", err)
	}

	g.Infof("  Removing %s", strings.TrimPrefix(filename, "../../../"))

	err = os.Remove(filename)
	if err != nil {
		g.Fatalf("error: %s", err)
	}
}

func readBadCaps(capsDataFile string) ([]string, error) {
	caps, err := common.ReadAllCSVData(capsDataFile)
	if err != nil {
		return nil, err
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

	slices.SortStableFunc(capsList, func(a, b string) int {
		return cmp.Or(
			// Reverse length order
			cmp.Compare(len(b), len(a)),
			cmp.Compare(a, b),
		)
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

	return strChunks, nil
}

func breakUpBigFile(g *common.Generator, filename, header string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	lines, err := lineCounter(f)
	if err != nil {
		return err
	}

	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(f)

	l := 0
	chunk := 0
	var w *bufio.Writer
	var piece *os.File
	var cfile string
	passedChunk := false

	re := regexache.MustCompile(`^  - id: `)

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
			g.Infof("  Splitting into %s", strings.TrimPrefix(cfile, "../../../"))
			chunk++

			var err error
			piece, err = os.OpenFile(cfile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return fmt.Errorf("opening file (%s): %s", cfile, err)
			}

			w = bufio.NewWriter(piece)
			_, err = w.WriteString(header)
			if err != nil {
				return fmt.Errorf("writing header to file (%s): %s", cfile, err)
			}
			w.Flush()
		}

		if w != nil {
			_, err = w.WriteString(fmt.Sprintf("%s\n", scanner.Text()))
			if err != nil {
				return fmt.Errorf("writing to file (%s): %s", cfile, err)
			}
		}

		l++
	}

	if w != nil {
		w.Flush()
	}

	return nil
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
