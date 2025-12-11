// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

//go:build ingore

package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"html/template"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/names/data"
	"golang.org/x/tools/go/packages"
)

// const (
// 	defaultFilename = "paginator_gen.go"
// )

var (
	// inputPaginator  = flag.String("InputPaginator", "", "name of the input pagination token field")
	listOp = flag.String("ListOp", "", "ListOp")
	// outputPaginator = flag.String("OutputPaginator", "", "name of the output pagination token field")
	// paginator       = flag.String("Paginator", "NextToken", "name of the pagination token field")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\tmain.go [flags] [<generated-lister-file>]\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	log.SetPrefix("generate/paginator: ")
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	// if (*inputPaginator != "" && *outputPaginator == "") || (*inputPaginator == "" && *outputPaginator != "") {
	// 	log.Fatal("both InputPaginator and OutputPaginator must be specified if one is")
	// }

	// if *inputPaginator == "" {
	// 	*inputPaginator = *paginator
	// }
	// if *outputPaginator == "" {
	// 	*outputPaginator = *paginator
	// }

	servicePackage := os.Getenv("GOPACKAGE")
	log.SetPrefix(fmt.Sprintf("generate/paginator: %s: ", servicePackage))

	service, err := data.LookupService(servicePackage)
	if err != nil {
		log.Fatalf("encountered: %s", err)
	}

	awsService := service.GoV2Package()

	function := *listOp

	tmpl := template.Must(template.New("function").Parse(functionTemplate))

	filename := fmt.Sprintf("paginator_%s_gen.go", function)
	if args := flag.Args(); len(args) > 0 {
		filename = args[0]
	}

	g := Generator{
		tmpl: tmpl,
		// inputPaginator:  *inputPaginator,
		// outputPaginator: *outputPaginator,
	}

	sourcePackage := fmt.Sprintf("github.com/aws/aws-sdk-go-v2/service/%[1]s", awsService)

	g.parsePackage(sourcePackage)

	g.printHeader(HeaderInfo{
		Parameters:         strings.Join(os.Args[1:], " "),
		DestinationPackage: servicePackage,
		SourcePackage:      sourcePackage,
	})

	// for _, functionName := range functions {
	// 	g.generateFunction(functionName, awsService, *export)
	// }
	g.generateFunction(function, awsService)

	src := g.format()

	err = os.WriteFile(filename, src, 0644)
	if err != nil {
		log.Fatalf("error writing output: %s", err)
	}
}

type HeaderInfo struct {
	Parameters         string
	DestinationPackage string
	SourcePackage      string
}

type Generator struct {
	buf  bytes.Buffer
	pkg  *Package
	tmpl *template.Template
	// inputPaginator  string
	// outputPaginator string
}

func (g *Generator) Printf(format string, args ...any) {
	fmt.Fprintf(&g.buf, format, args...)
}

type PackageFile struct {
	file *ast.File
}

type Package struct {
	name  string
	files []*PackageFile
}

func (g *Generator) printHeader(headerInfo HeaderInfo) {
	header := template.Must(template.New("header").Parse(headerTemplate))

	err := header.Execute(&g.buf, headerInfo)
	if err != nil {
		log.Fatalf("error writing header: %s", err)
	}
}

func (g *Generator) parsePackage(sourcePackage string) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedSyntax,
	}
	pkgs, err := packages.Load(cfg, sourcePackage)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages found", len(pkgs))
	}
	g.addPackage(pkgs[0])
}

func (g *Generator) addPackage(pkg *packages.Package) {
	g.pkg = &Package{
		name:  pkg.Name,
		files: make([]*PackageFile, len(pkg.Syntax)),
	}

	for i, file := range pkg.Syntax {
		g.pkg.files[i] = &PackageFile{
			file: file,
		}
	}
}

type FuncSpec struct {
	Name      string
	AWSName   string
	NoSemgrep string
	// AWSService string
	// ParamType  string
	// ResultType string
	// InputPaginator  string
	// OutputPaginator string
}

func (g *Generator) generateFunction(functionName, awsService string) {
	var function *ast.FuncDecl

	for _, file := range g.pkg.files {
		if file.file != nil {
			for _, decl := range file.file.Decls {
				if funcDecl, ok := decl.(*ast.FuncDecl); ok {
					if funcDecl.Name.Name == functionName {
						function = funcDecl
						break
					}
				}
			}
			if function != nil {
				break
			}
		}
	}

	if function == nil {
		log.Fatalf("function \"%s\" not found", functionName)
	}

	funcName := function.Name.Name

	// if !export {
	funcName = fmt.Sprintf("%s%s", strings.ToLower(funcName[0:1]), funcName[1:])
	// }

	funcSpec := FuncSpec{
		Name:      funcName,
		AWSName:   function.Name.Name,
		NoSemgrep: fixSomeInitialisms(funcName),
		// AWSService: awsService,
		// ParamType:  g.expandTypeField(function.Type.Params, false), // Assumes there is a single input parameter
		// ResultType: g.expandTypeField(function.Type.Results, true), // Assumes we can take the first return parameter
		// InputPaginator:  g.inputPaginator,
		// OutputPaginator: g.outputPaginator,
	}

	err := g.tmpl.Execute(&g.buf, funcSpec)
	if err != nil {
		log.Fatalf("error writing function \"%s\": %s", functionName, err)
	}
}

func (g *Generator) expandTypeField(field *ast.FieldList, result bool) string {
	typeValue := field.List[0].Type

	if !result {
		typeValue = field.List[1].Type
	}

	if star, ok := typeValue.(*ast.StarExpr); ok {
		return fmt.Sprintf("*%s", g.expandTypeExpr(star.X))
	}

	log.Fatalf("Unexpected type expression: (%[1]T) %[1]v", typeValue)
	return ""
}

func (g *Generator) expandTypeExpr(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return fmt.Sprintf("%s.%s", g.pkg.name, ident.Name)
	}

	log.Fatalf("Unexpected expression: (%[1]T) %[1]v", expr)
	return ""
}

//go:embed header.gtpl
var headerTemplate string

//go:embed function.gtpl
var functionTemplate string

func (g *Generator) format() []byte {
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return g.buf.Bytes()
	}
	return src
}

func fixSomeInitialisms(s string) string {
	if strings.Contains(s, "Vpc") {
		return "caps5"
	}
	return ""
}
