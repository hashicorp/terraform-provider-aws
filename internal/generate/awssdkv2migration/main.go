package main

import (
	_ "embed"
	"flag"
	"fmt"
	"go/ast"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names/data"
	"golang.org/x/tools/go/packages"
)

const (
	relativePath = "../../service"
	filename     = "aws_sdk_v2.patch"
)

type TemplateData struct {
	GoV1Package        string
	GoV1ClientTypeName string
	GoV2Package        string
	ProviderPackage    string
	InputOutputTypes   []string
	ContextFunctions   []string
	Exceptions         []string
}

//go:embed patch.tmpl
var tmpl string

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\tmain.go [flags]\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func getPackageData(goV1Package string, goV1Client string, goV2Package string, providerPackage string) TemplateData {
	cwd, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	config := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedFiles,
		Dir:  path.Join(cwd, "..", "..", "service", providerPackage),
	}

	var td = TemplateData{goV1Package, goV1Client, goV2Package, providerPackage, []string{}, []string{}, []string{}}

	pkgs, err := packages.Load(config, "github.com/aws/aws-sdk-go/service/"+goV1Package)

	if err != nil {
		fmt.Printf("Error loading package %s: %s", goV1Package, err)
	}

	for _, s := range pkgs[0].Syntax {
		for n, o := range s.Scope.Objects {
			if o.Kind == ast.Typ && (strings.HasSuffix(n, "Input") || strings.HasSuffix(n, "Output")) {
				td.InputOutputTypes = append(td.InputOutputTypes, n)
			}

			if o.Kind == ast.Con && strings.HasPrefix(n, "op") {
				td.ContextFunctions = append(td.ContextFunctions, strings.TrimPrefix(n, "op"))
			}

			if o.Kind == ast.Typ && strings.HasSuffix(n, "Exception") {
				td.Exceptions = append(td.Exceptions, strings.TrimPrefix(n, "ErrCode"))
			}
		}
	}

	return td
}

func main() {
	log.SetPrefix("generate/sdkv2migration: ")
	log.SetFlags(0)

	g := common.NewGenerator()

	data, err := data.ReadAllServiceData()

	if err != nil {
		g.Fatalf("error reading service data: %s", err)
	}

	for _, l := range data {
		if l.Exclude() {
			continue
		}

		if l.NotImplemented() && !l.EndpointOnly() {
			continue
		}

		packageName := l.ProviderPackage()
		d := g.NewUnformattedFileDestination(filepath.Join(relativePath, packageName, filename))
		td := getPackageData(l.GoV1Package(), l.GoV1ClientTypeName(), l.GoV2Package(), packageName)

		if err := d.WriteTemplate("awssdkv2migration", tmpl, td); err != nil {
			g.Fatalf("error generating %s aws sdk v2 migration patch: %s", packageName, err)
		}

		if err := d.Write(); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}
	}
}
