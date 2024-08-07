// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	_ "embed"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
)

//go:embed consts.tmpl
var constsTmpl string

//go:embed constOrQuote.gtpl
var constOrQuoteTmpl string

//go:embed semgrep.tmpl
var semgrepTmpl string

type ConstantDatum struct {
	Constant   string
	Literal    string
	AltLiteral string
}

type TemplateData struct {
	Constants []ConstantDatum
}

func main() {
	const (
		constsFilename       = "../../../internal/acctest/consts_gen.go"
		constOrQuoteFilename = "../../../internal/acctest/generate/const_or_quote_gen.go"
		semgrepFilename      = "../../../.ci/.semgrep-test-constants.yml"
		constantDataFile     = "../../../internal/acctest/consts.csv"
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

	if err := d.WriteTemplate("constantlist", constsTmpl, td); err != nil {
		g.Fatalf("generating file (%s): %s", constsFilename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", constsFilename, err)
	}

	// ConstOrQuote helper
	g.Infof("Generating %s", strings.TrimPrefix(constOrQuoteFilename, "../../../"))

	d = g.NewGoFileDestination(constOrQuoteFilename)

	if err := d.WriteTemplate("constsOrQuote", constOrQuoteTmpl, td); err != nil {
		g.Fatalf("generating file (%s): %s", constOrQuoteFilename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", constOrQuoteFilename, err)
	}

	// Semgrep
	g.Infof("Generating %s", strings.TrimPrefix(semgrepFilename, "../../../"))

	d = g.NewUnformattedFileDestination(semgrepFilename)

	if err := d.WriteTemplate("semgrep-constants", semgrepTmpl, td); err != nil {
		g.Fatalf("generating file (%s): %s", semgrepFilename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", semgrepFilename, err)
	}
}

func readConstants(filename string) ([]ConstantDatum, error) {
	constants, err := common.ReadAllCSVData(filename)

	if err != nil {
		return nil, err
	}

	var constantList []ConstantDatum

	for _, row := range constants {
		if row[0] == "" {
			continue
		}

		cd := ConstantDatum{
			Literal:  row[0],
			Constant: row[1],
		}

		switch cd.Literal {
		case "0":
			cd.AltLiteral = "acctest.CtZero"
		case "1":
			cd.AltLiteral = "acctest.CtOne"
		case "2":
			cd.AltLiteral = "acctest.CtTwo"
		case "3":
			cd.AltLiteral = "acctest.CtThree"
		case "4":
			cd.AltLiteral = "acctest.CtFour"
		case "10":
			cd.AltLiteral = "acctest.CtTen"
		}

		constantList = append(constantList, cd)
	}

	sort.SliceStable(constantList, func(i, j int) bool {
		return constantList[j].Constant > constantList[i].Constant
	})

	return constantList, nil
}
