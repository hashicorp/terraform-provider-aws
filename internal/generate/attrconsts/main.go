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

// names/attr_consts_gen.go is used by other generators so it is very important
// that it is generated first. This is accomplished by using generate.go in this
// directory rather names/generate.go.

//go:embed file.tmpl
var tmpl string

//go:embed semgrep.tmpl
var semgrepTmpl string

type ConstantDatum struct {
	Constant      string
	ConstantLower string
}

type TemplateData struct {
	Constants []ConstantDatum
}

func main() {
	const (
		filename         = "../../../names/attr_consts_gen.go"
		semgrepFilename  = "../../../.ci/.semgrep-constants.yml"
		constantDataFile = "../../../names/attr_constants.csv"
	)
	g := common.NewGenerator()

	constants, err := readConstants(constantDataFile)

	if err != nil {
		g.Fatalf("error reading %s: %s", constantDataFile, err)
	}

	td := TemplateData{}
	td.Constants = constants

	g.Infof("Generating %s", strings.TrimPrefix(filename, "../../../"))

	d := g.NewGoFileDestination(filename)

	if err := d.WriteTemplate("constantlist", tmpl, td); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

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

		constantList = append(constantList, ConstantDatum{
			ConstantLower: row[0],
			Constant:      row[1],
		})
	}

	sort.SliceStable(constantList, func(i, j int) bool {
		return constantList[j].Constant > constantList[i].Constant
	})

	return constantList, nil
}
