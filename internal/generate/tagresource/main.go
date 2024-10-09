// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names"
	namesgen "github.com/hashicorp/terraform-provider-aws/names/generate"
)

var (
	getTagFunc     = flag.String("GetTagFunc", "findTag", "getTagFunc")
	idAttribName   = flag.String("IDAttribName", "resource_arn", "idAttribName")
	updateTagsFunc = flag.String("UpdateTagsFunc", "updateTags", "updateTagsFunc")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\tmain.go [flags]\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

type TemplateData struct {
	AWSServiceUpper      string
	ProviderResourceName string
	ServicePackage       string

	GetTagFunc     string
	IDAttribName   string
	UpdateTagsFunc string
}

func main() {
	const (
		resourceFilename     = `tag_gen.go`
		resourceTestFilename = `tag_gen_test.go`
	)
	g := common.NewGenerator()

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	servicePackage := os.Getenv("GOPACKAGE")
	u, err := names.ProviderNameUpper(servicePackage)
	if err != nil {
		g.Fatalf("encountered: %s", err)
	}

	resourceName := fmt.Sprintf("aws_%s_tag", servicePackage)

	ian := *idAttribName
	ian = namesgen.ConstOrQuote(ian)

	templateData := TemplateData{
		AWSServiceUpper:      u,
		ProviderResourceName: resourceName,
		ServicePackage:       servicePackage,

		GetTagFunc:     *getTagFunc,
		IDAttribName:   ian,
		UpdateTagsFunc: *updateTagsFunc,
	}

	g.Infof("Generating internal/service/%s/%s", servicePackage, resourceFilename)
	d := g.NewGoFileDestination(resourceFilename)

	if err := d.WriteTemplate("taggen", resourceTemplateBody, templateData); err != nil {
		g.Fatalf("generating file (%s): %s", resourceFilename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", resourceFilename, err)
	}

	g.Infof("Generating internal/service/%s/%s", servicePackage, resourceTestFilename)
	d = g.NewGoFileDestination(resourceTestFilename)

	if err := d.WriteTemplate("taggen", resourceTestTemplateBody, templateData); err != nil {
		g.Fatalf("generating file (%s): %s", resourceTestFilename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", resourceTestFilename, err)
	}
}

//go:embed resource.tmpl
var resourceTemplateBody string

//go:embed tests.tmpl
var resourceTestTemplateBody string
