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
	createTagsFunc = flag.String("CreateTagsFunc", "CreateTags", "createTagsFunc")
	getTagFunc     = flag.String("GetTagFunc", "GetTag", "getTagFunc")
	idAttribName   = flag.String("IDAttribName", "resource_arn", "idAttribName")
	updateTagsFunc = flag.String("UpdateTagsFunc", "UpdateTags", "updateTagsFunc")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\tmain.go [flags]\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

type TemplateData struct {
	AWSService      string
	AWSServiceUpper string
	ServicePackage  string

	CreateTagsFunc string
	GetTagFunc     string
	IDAttribName   string
	UpdateTagsFunc string
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

	templateData := TemplateData{
		AWSService:      awsService,
		AWSServiceUpper: u,
		ServicePackage:  servicePackage,

		CreateTagsFunc: *createTagsFunc,
		GetTagFunc:     *getTagFunc,
		IDAttribName:   *idAttribName,
		UpdateTagsFunc: *updateTagsFunc,
	}

	resourceFilename := "tag_gen.go"
	d := g.NewGoFileDestination(resourceFilename)

	if err := d.WriteTemplate("taggen", resourceTemplateBody, templateData); err != nil {
		g.Fatalf("error: %s", err.Error())
	}

	resourceTestFilename := "tag_gen_test.go"
	d = g.NewGoFileDestination(resourceTestFilename)

	if err := d.WriteTemplate("taggen", resourceTestTemplateBody, templateData); err != nil {
		g.Fatalf("error: %s", err.Error())
	}
}

//go:embed resource.tmpl
var resourceTemplateBody string

//go:embed tests.tmpl
var resourceTestTemplateBody string
