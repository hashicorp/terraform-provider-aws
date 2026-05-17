// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

//go:build generate

package main

import (
	_ "embed"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
)

//go:embed lookup.tmpl
var tmpl string

type TagType struct {
	Name           string   `hcl:"name,label"`
	TerraformTypes []string `hcl:"terraform_types"`
}

type Config struct {
	TagTypes []TagType `hcl:"tagtype,block"`
}

func main() {
	const (
		source   = `../../../internal/tags/tagpolicy/tagris-terraform-mapping.hcl`
		filename = `../../../internal/tags/tagpolicy/lookup_gen.go`
	)
	g := common.NewGenerator()

	g.Infof("Generating %s", strings.TrimPrefix(filename, "../../../"))

	var config Config
	err := hclsimple.DecodeFile(source, nil, &config)
	if err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	// Sort by tag resource name
	slices.SortFunc(config.TagTypes, func(a, b TagType) int {
		return strings.Compare(a.Name, b.Name)
	})

	data := map[string]any{
		"TagTypes": config.TagTypes,
	}

	d := g.NewGoFileDestination(filename)

	if err := d.BufferTemplate("lookup", tmpl, data); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}
