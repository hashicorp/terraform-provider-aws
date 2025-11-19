// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate

package main

import (
	"cmp"
	_ "embed"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
)

//go:embed lookup.tmpl
var tmpl string

func main() {
	const (
		source   = `../../../internal/tags/tagpolicy/tagris-cfn-terraform-mapping.csv`
		filename = `../../../internal/tags/tagpolicy/lookup_gen.go`
	)
	g := common.NewGenerator()

	g.Infof("Generating %s", strings.TrimPrefix(filename, "../../../"))

	rows, err := common.ReadAllCSVData(source)
	if err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	type m struct {
		Tagris string
		Tf     string
	}
	var mapping []m
	for _, r := range rows[1:] {
		if len(r) != 3 {
			continue
		}

		if r[2] != "" {
			mapping = append(mapping, m{Tagris: r[1], Tf: r[2]})
			continue
		}

		// g.Infof("No lookup match found for: %s", r[1])
	}

	// Sort by tag type name
	slices.SortFunc(mapping, func(a, b m) int {
		return cmp.Compare(a.Tagris, b.Tagris)
	})

	data := map[string]any{
		"Mapping": mapping,
	}

	d := g.NewGoFileDestination(filename)

	if err := d.BufferTemplate("lookup", tmpl, data); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}
