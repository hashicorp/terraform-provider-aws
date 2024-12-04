// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	"cmp"
	_ "embed"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
)

//go:embed header.tmpl
var header string

//go:embed file.tmpl
var tmpl string

const (
	maxBadCaps = 31
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
	const (
		filename     = "../../../names/caps.md"
		capsDataFile = "../../../names/caps.csv"
	)
	g := common.NewGenerator()

	g.Infof("Generating %s", strings.TrimPrefix(filename, "../../../"))

	badCaps, err := readBadCaps(capsDataFile)

	if err != nil {
		g.Fatalf("error reading %s: %s", capsDataFile, err)
	}

	td := TemplateData{}
	td.BadCaps = badCaps

	d := g.NewUnformattedFileDestination(filename)

	if err := d.BufferTemplate("namescapslist", header+"\n"+tmpl+"\n", td); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

func readBadCaps(filename string) ([]CapsDatum, error) {
	caps, err := common.ReadAllCSVData(filename)

	if err != nil {
		return nil, err
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

	slices.SortStableFunc(capsList, func(a, b CapsDatum) int {
		return cmp.Or(
			// Reverse length order
			cmp.Compare(len(b.Wrong), len(a.Wrong)),
			cmp.Compare(a.Wrong, b.Wrong),
		)
	})

	onChunk := -1

	for i := range capsList {
		if i%maxBadCaps == 0 {
			onChunk++
		}

		capsList[i].Test = fmt.Sprintf(`%s%d`, "caps", onChunk)
	}

	slices.SortStableFunc(capsList, func(a, b CapsDatum) int {
		return cmp.Or(
			cmp.Compare(strings.ToLower(a.Wrong), strings.ToLower(b.Wrong)),
			cmp.Compare(a.Wrong, b.Wrong),
		)
	})

	return capsList, nil
}
