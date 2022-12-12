//go:build generate
// +build generate

package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\tmain.go [flags] [<generated-file>]\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	filename := `service_package_data_gen.go`
	if args := flag.Args(); len(args) > 0 {
		filename = args[0]
	}

	g := common.NewGenerator()
	templateData := &TemplateData{
		PackageName: os.Getenv("GOPACKAGE"),
	}

	if err := g.ApplyAndWriteGoTemplate(filename, "servicepackagedata", tmpl, templateData); err != nil {
		g.Fatalf("error generating %s service package data: %s", templateData.PackageName, err.Error())
	}
}

type TemplateData struct {
	PackageName string
}

//go:embed file.tmpl
var tmpl string
