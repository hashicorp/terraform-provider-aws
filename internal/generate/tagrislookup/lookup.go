// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	_ "embed"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
)

//go:embed lookup.tmpl
var tmpl string

func main() {
	const (
		source   = `../../../internal/tags/tagris/tagris-cfn-terraform-mapping.csv`
		filename = `../../../internal/tags/tagris/lookup.go`
	)
	g := common.NewGenerator()

	g.Infof("Generating %s", strings.TrimPrefix(filename, "../../../"))

	rows, err := common.ReadAllCSVData(source)
	if err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	mapping := make(map[string]string, len(rows))
	for _, r := range rows[1:] {
		if len(r) != 3 {
			continue
		}

		if r[2] != "" {
			mapping[r[1]] = r[2]
			continue
		}

		// g.Infof("No lookup match found for: %s", r[1])
	}

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
