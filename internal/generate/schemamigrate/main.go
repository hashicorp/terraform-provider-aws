//go:build ignore
// +build ignore

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/cli"
)

var (
	tfResourceType = flag.String("resource", "", "Terraform resource type")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\tmain.go [flags] -resource <TF-resource-type> <generated-schema-file>\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if len(args) < 2 || *tfResourceType == "" {
		flag.Usage()
		os.Exit(2)
	}

	outputFilename := args[0]

	ui := &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}
	generator := &schemaGenerator{
		ui: ui,
	}

	if err := generator.generate(outputFilename); err != nil {
		ui.Error(fmt.Sprintf("error migrating Terraform %s schema: %s", *tfResourceType, err))
		os.Exit(1)
	}
}

type schemaGenerator struct {
	sdkSchema map[string]*schema.Schema
	ui        cli.Ui
}

func (g *schemaGenerator) generate(outputFilename string) error {
	return nil
}
