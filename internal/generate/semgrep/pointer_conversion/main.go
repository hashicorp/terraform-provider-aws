// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

//go:build generate

package main

import (
	_ "embed"
	"strings"
	"text/template"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
)

type typ struct {
	Typ          string
	WrapFunc     string
	UnwrapFunc   string
	Conditionals []string
}

func main() {
	const (
		rulesFilename   = `../../../../.ci/semgrep/aws/pointer-conversion.yml`
		testFilename    = `../../../../.ci/semgrep/aws/pointer-conversion.go`
		fixTestFilename = `../../../../.ci/semgrep/aws/pointer-conversion.fixed.go`
	)

	types := map[string]typ{
		"bool": {
			WrapFunc:   "aws.Bool",
			UnwrapFunc: "aws.ToBool",
		},
		"float32": {
			WrapFunc:   "aws.Float32",
			UnwrapFunc: "aws.ToFloat32",
		},
		"float64": {
			WrapFunc:   "aws.Float64",
			UnwrapFunc: "aws.ToFloat64",
		},
		"int32": {
			WrapFunc:   "aws.Int32",
			UnwrapFunc: "aws.ToInt32",
		},
		"int64": {
			WrapFunc:   "aws.Int64",
			UnwrapFunc: "aws.ToInt64",
		},
		"string": {
			WrapFunc:   "aws.String",
			UnwrapFunc: "aws.ToString",
		},
		"time": {
			Typ:        "time.Time",
			WrapFunc:   "aws.Time",
			UnwrapFunc: "aws.ToTime",
		},
	}
	for k := range types {
		typ := types[k]
		if k == "bool" {
			typ.Conditionals = []string{"==", "!="}
		} else {
			typ.Conditionals = []string{"==", "!=", ">", "<", ">=", "<="}
		}
		if typ.Typ == "" {
			typ.Typ = k
		}
		types[k] = typ
	}

	g := common.NewGenerator()

	d := g.NewUnformattedFileDestination(rulesFilename)

	if err := d.BufferTemplate("rules", rulesTemplate, types); err != nil {
		g.Fatalf("generating file (%s): %s", rulesFilename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", rulesFilename, err)
	}

	funcs := template.FuncMap{
		"Title": strings.Title,
	}

	d = g.NewUnformattedFileDestination(testFilename)

	if err := d.BufferTemplate("tests", testTemplate, types, funcs); err != nil {
		g.Fatalf("generating file (%s): %s", testFilename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", testFilename, err)
	}

	d = g.NewGoFileDestination(fixTestFilename)

	if err := d.BufferTemplate("fix-tests", fixTestTemplate, types, funcs); err != nil {
		g.Fatalf("generating file (%s): %s", fixTestFilename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", fixTestFilename, err)
	}
}

//go:embed rules.tmpl
var rulesTemplate string

//go:embed test.tmpl
var testTemplate string

//go:embed fix-test.tmpl
var fixTestTemplate string
