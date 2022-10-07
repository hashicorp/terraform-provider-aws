//go:build generate
// +build generate

package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"text/template"

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
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	servicePackage := os.Getenv("GOPACKAGE")
	awsService, err := names.AWSGoV1Package(servicePackage)

	if err != nil {
		log.Fatalf("encountered: %s", err)
	}

	u, err := names.ProviderNameUpper(servicePackage)

	if err != nil {
		log.Fatalf("encountered: %s", err)
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
	resourceTestFilename := "tag_gen_test.go"

	if err := generateTemplateFile(resourceFilename, resourceTemplateBody, templateData); err != nil {
		log.Fatal(err)
	}

	if err := generateTemplateFile(resourceTestFilename, resourceTestTemplateBody, templateData); err != nil {
		log.Fatal(err)
	}
}

func generateTemplateFile(filename string, templateBody string, templateData interface{}) error {
	tmpl, err := template.New(filename).Parse(templateBody)

	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, templateData)

	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	generatedFileContents, err := format.Source(buffer.Bytes())

	if err != nil {
		return fmt.Errorf("error formatting generated file: %w", err)
	}

	f, err := os.Create(filename)

	if err != nil {
		return fmt.Errorf("error creating file (%s): %w", filename, err)
	}

	defer f.Close()

	_, err = f.Write(generatedFileContents)

	if err != nil {
		return fmt.Errorf("error writing to file (%s): %w", filename, err)
	}

	return nil
}

//go:embed resource.tmpl
var resourceTemplateBody string

//go:embed tests.tmpl
var resourceTestTemplateBody string
