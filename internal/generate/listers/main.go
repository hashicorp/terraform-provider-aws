// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

//go:build generate

package main

import (
	_ "embed"
	"flag"
	"fmt"
	"go/types"
	"os"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names/data"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\tmain.go [flags] [<generated-lister-file>]\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	g := common.NewGenerator()

	filename := "list_gen.go"
	if args := flag.Args(); len(args) > 0 {
		filename = args[0]
	}

	// Finder functions.
	functionNameRegex := regexache.MustCompile(`^(?:Describe|Get)((?:\w+)s)$`)
	// TODO Consider the RESTful services, e.g. apigatewayv2 with GetApi and GetApis returning different types.

	servicePackage := os.Getenv("GOPACKAGE")

	g.Infof("Generating internal/service/%s/%s", servicePackage, filename)

	service, err := data.LookupService(servicePackage)
	if err != nil {
		g.Fatalf("encountered: %s", err)
	}

	awsPkg := service.GoPackageName()

	sourcePackage := fmt.Sprintf("github.com/aws/aws-sdk-go-v2/service/%[1]s", awsPkg)
	pkg, err := common.LoadPackage(sourcePackage)
	if err != nil {
		g.Fatalf("parsing package (%s): %s", sourcePackage, err)
	}

	// Functions are on the 'Client' type.
	scope := pkg.Scope()
	obj := scope.Lookup("Client")
	if obj == nil {
		g.Fatalf("Client type not found in package scope")
	}

	named := obj.(*types.TypeName).Type().(*types.Named)
	for i := range named.NumMethods() {
		fn := named.Method(i)
		functionName := fn.Name()
		match := functionNameRegex.FindStringSubmatch(functionName)
		if len(match) <= 1 {
			continue
		}
		pluralName := match[1]

		// Is the output paginated?
		var isPaginated bool
		var singleTypeName string
		structName := fmt.Sprintf("%sOutput", functionName)
		obj := scope.Lookup(structName)
		if typeName, ok := obj.(*types.TypeName); ok {
			if named, ok := typeName.Type().(*types.Named); ok {
				if s, ok := named.Underlying().(*types.Struct); ok {
					for i := range s.NumFields() {
						f := s.Field(i)
						if fieldName := f.Name(); fieldName == "NextToken" {
							isPaginated = true
						} else if fieldName == pluralName {
							if slice, ok := f.Type().(*types.Slice); ok {
								singleTypeName = types.TypeString(slice.Elem(), func(*types.Package) string { return "" })
							}
						}
					}
				}
			}
		}

		// Look for a paginator.
		var hasPaginator bool
		paginatorName := fmt.Sprintf("New%sPaginator", functionName)
		obj = scope.Lookup(paginatorName)
		if _, ok := obj.(*types.Func); ok {
			hasPaginator = true
		}

		g.Infof("  %s, type: %s, paginated: %t, paginator: %t", functionName, singleTypeName, isPaginated, hasPaginator)
	}
}
