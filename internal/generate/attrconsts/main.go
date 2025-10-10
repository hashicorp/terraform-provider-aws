// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	"cmp"
	_ "embed"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
)

// names/attr_consts_gen.go is used by other generators so it is very important
// that it is generated first. This is accomplished by using generate.go in this
// directory rather names/generate.go.

//go:embed consts.gtpl
var tmpl string

//go:embed constOrQuote.gtpl
var constOrQuoteTmpl string

//go:embed semgrep.gtpl
var semgrepTmpl string

type constantDatum struct {
	Constant string
	Literal  string
}

type TemplateData struct {
	Constants []constantDatum
}

func main() {
	const (
		constsFilename       = "../../../names/attr_consts_gen.go"
		constOrQuoteFilename = "../../../names/generate/const_or_quote_gen.go"
		semgrepFilename      = "../../../.ci/.semgrep-constants.yml"
		constantDataFile     = "../../../names/attr_constants.csv"
	)
	g := common.NewGenerator()

	constants, err := readConstants(constantDataFile)

	if err != nil {
		g.Fatalf("error reading %s: %s", constantDataFile, err)
	}

	td := TemplateData{}
	td.Constants = constants

	// Constants file
	g.Infof("Generating %s", strings.TrimPrefix(constsFilename, "../../../"))

	d := g.NewGoFileDestination(constsFilename)

	if err := d.BufferTemplate("constantlist", tmpl, td); err != nil {
		g.Fatalf("generating file (%s): %s", constsFilename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", constsFilename, err)
	}

	// ConstsOrQuotes helper
	g.Infof("Generating %s", strings.TrimPrefix(constOrQuoteFilename, "../../../"))

	d = g.NewGoFileDestination(constOrQuoteFilename)

	if err := d.BufferTemplate("constOrQuote", constOrQuoteTmpl, td); err != nil {
		g.Fatalf("generating file (%s): %s", constOrQuoteFilename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", constOrQuoteFilename, err)
	}

	// Semgrep
	g.Infof("Generating %s", strings.TrimPrefix(semgrepFilename, "../../../"))

	d = g.NewUnformattedFileDestination(semgrepFilename)

	if err := d.BufferTemplate("semgrep-constants", semgrepTmpl, td); err != nil {
		g.Fatalf("generating file (%s): %s", semgrepFilename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", semgrepFilename, err)
	}
}

func readConstants(filename string) ([]constantDatum, error) {
	constants, err := common.ReadAllCSVData(filename)

	if err != nil {
		return nil, err
	}

	var constantList []constantDatum

	for _, row := range constants {
		if row[0] == "" {
			continue
		}

		constantList = append(constantList, constantDatum{
			Literal:  row[0],
			Constant: row[1],
		})
	}

	slices.SortStableFunc(constantList, func(a, b constantDatum) int {
		return cmp.Compare(a.Constant, b.Constant)
	})

	return constantList, nil
}
