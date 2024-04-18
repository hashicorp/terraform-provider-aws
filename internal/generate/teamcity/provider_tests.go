// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
)

type ServiceDatum struct {
	ProviderPackage string
	HumanFriendly   string
	VpcLock         bool
	Parallelism     int
	Region          string
}

type TemplateData struct {
	Services []ServiceDatum
}

func main() {
	const (
		acceptanceTestsScriptFile = `.teamcity/scripts/provider_tests/acceptance_tests.sh`
		unitTestsScriptFile       = `.teamcity/scripts/provider_tests/unit_tests.sh`
	)
	g := common.NewGenerator()

	projectRoot, err := filepath.Abs(`../../../`)
	if err != nil {
		g.Fatalf(err.Error())
	}

	internalDir := filepath.Join(projectRoot, "internal")

	dirs, err := os.ReadDir(internalDir)
	if err != nil {
		g.Fatalf(err.Error())
	}

	generator := generator{
		g:    g,
		root: projectRoot,
	}

	for _, dir := range dirs {
		if dir.IsDir() && dir.Name() != "service" {
			generator.dirNames = append(generator.dirNames, dir.Name())
		}
	}

	generator.generate(acceptanceTestsScriptFile, acceptanceTestsTmpl)

	generator.generate(unitTestsScriptFile, unitTestsTmpl)
}

type generator struct {
	g        *common.Generator
	root     string
	dirNames []string
}

func (g generator) generate(filename, template string) {
	g.g.Infof("Generating %s", filename)

	destFile := filepath.Join(g.root, filename)

	d := g.g.NewUnformattedFileDestination(destFile)

	if err := d.WriteTemplate("teamcity", template, g.dirNames); err != nil {
		g.g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.g.Fatalf("generating file (%s): %s", filename, err)
	}
}

//go:embed acceptance_tests.tmpl
var acceptanceTestsTmpl string

//go:embed unit_tests.tmpl
var unitTestsTmpl string
