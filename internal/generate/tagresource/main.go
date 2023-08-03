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
)

var (
	getTagFunc     = flag.String("GetTagFunc", "GetTag", "getTagFunc")
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
	AWSService           string
	AWSServiceUpper      string
	ProviderResourceName string
	ServicePackage       string

	CreateTagsFunc string
	GetTagFunc     string
	IDAttribName   string

	UpdateTagsFunc string
	WithContext    bool
}

func main() {
	g := common.NewGenerator()

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	servicePackage := os.Getenv("GOPACKAGE")
	awsService, err := names.AWSGoV1Package(servicePackage)

	if err != nil {
		g.Fatalf("encountered: %s", err)
	}

	u, err := names.ProviderNameUpper(servicePackage)

	if err != nil {
		g.Fatalf("encountered: %s", err)
	}

	providerResName := fmt.Sprintf("aws_%s_tag", servicePackage)

	templateData := TemplateData{
		AWSService:           awsService,
		AWSServiceUpper:      u,
		ProviderResourceName: providerResName,
		ServicePackage:       servicePackage,

		GetTagFunc:     *getTagFunc,
		IDAttribName:   *idAttribName,
		UpdateTagsFunc: *updateTagsFunc,
	}

	const (
		resourceFilename     = "tag_gen.go"
		resourceTestFilename = "tag_gen_test.go"
	)

	d := g.NewGoFileDestination(resourceFilename)

	if err := d.WriteTemplate("taggen", resourceTemplateBody, templateData); err != nil {
		g.Fatalf("generating file (%s): %s", resourceFilename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", resourceFilename, err)
	}

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
