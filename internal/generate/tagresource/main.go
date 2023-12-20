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

const (
	sdkV1 = 1
	sdkV2 = 2
)

var (
	getTagFunc     = flag.String("GetTagFunc", "GetTag", "getTagFunc")
	idAttribName   = flag.String("IDAttribName", "resource_arn", "idAttribName")
	sdkVersion     = flag.Int("AWSSDKVersion", sdkV1, "Version of the AWS Go SDK to use i.e. 1 or 2")
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

	if *sdkVersion != sdkV1 && *sdkVersion != sdkV2 {
		g.Fatalf("AWS SDK Go Version %d not supported", *sdkVersion)
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

	var resourceTemplateBody string
	var resourceTestTemplateBody string

	switch *sdkVersion {
	case sdkV1:
		resourceTemplateBody = v1ResourceTemplateBody
		resourceTestTemplateBody = v1ResourceTestTemplateBody
	case sdkV2:
		resourceTemplateBody = v2ResourceTemplateBody
		resourceTestTemplateBody = v2ResourceTestTemplateBody
	}

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

//go:embed templates/v1/resource.tmpl
var v1ResourceTemplateBody string

//go:embed templates/v2/resource.tmpl
var v2ResourceTemplateBody string

//go:embed templates/v1/tests.tmpl
var v1ResourceTestTemplateBody string

//go:embed templates/v2/tests.tmpl
var v2ResourceTestTemplateBody string
