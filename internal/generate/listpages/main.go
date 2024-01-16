// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

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
	"sort"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/names"
	"golang.org/x/tools/go/packages"
)

const (
	defaultFilename = "list_pages_gen.go"
)

var (
	inputPaginator  = flag.String("InputPaginator", "", "name of the input pagination token field")
	listOps         = flag.String("ListOps", "", "ListOps")
	outputPaginator = flag.String("OutputPaginator", "", "name of the output pagination token field")
	paginator       = flag.String("Paginator", "NextToken", "name of the pagination token field")
	export          = flag.Bool("Export", false, "whether to export the list functions")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\tmain.go [flags] [<generated-lister-file>]\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	log.SetPrefix("generate/listpage: ")
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	if (*inputPaginator != "" && *outputPaginator == "") || (*inputPaginator == "" && *outputPaginator != "") {
		log.Fatal("both InputPaginator and OutputPaginator must be specified if one is")
	}

	if *inputPaginator == "" {
		*inputPaginator = *paginator
	}
	if *outputPaginator == "" {
		*outputPaginator = *paginator
	}

	filename := defaultFilename
	if args := flag.Args(); len(args) > 0 {
		filename = args[0]
	}

	servicePackage := os.Getenv("GOPACKAGE")
	log.SetPrefix(fmt.Sprintf("generate/listpage: %s: ", servicePackage))

	awsService, err := names.AWSGoV1Package(servicePackage)

	if err != nil {
		log.Fatalf("encountered: %s", err)
	}

	functions := strings.Split(*listOps, ",")
	sort.Strings(functions)

	g := Generator{
		tmpl:            template.Must(template.New("function").Parse(functionTemplate)),
		inputPaginator:  *inputPaginator,
		outputPaginator: *outputPaginator,
	}

	sourcePackage := fmt.Sprintf("github.com/aws/aws-sdk-go/service/%[1]s", awsService)
	g.parsePackage(sourcePackage)

	g.printHeader(HeaderInfo{
		Parameters:         strings.Join(os.Args[1:], " "),
		DestinationPackage: servicePackage,
		SourcePackage:      sourcePackage,
		SourceIntfPackage:  fmt.Sprintf("github.com/aws/aws-sdk-go/service/%[1]s/%[1]siface", awsService),
	})

	awsUpper, err := names.AWSGoV1ClientTypeName(servicePackage)

	if err != nil {
		log.Fatalf("encountered: %s", err)
	}

	for _, functionName := range functions {
		g.generateFunction(functionName, awsService, awsUpper, *export)
	}

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
	SourceIntfPackage  string
}

type Generator struct {
	buf             bytes.Buffer
	pkg             *Package
	tmpl            *template.Template
	inputPaginator  string
	outputPaginator string
}

func (g *Generator) Printf(format string, args ...interface{}) {
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
	Name            string
	AWSName         string
	RecvType        string
	ParamType       string
	ResultType      string
	InputPaginator  string
	OutputPaginator string
}

func (g *Generator) generateFunction(functionName, awsService, awsServiceUpper string, export bool) {
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

	if !export {
		funcName = fmt.Sprintf("%s%s", strings.ToLower(funcName[0:1]), funcName[1:])
	}

	funcSpec := FuncSpec{
		Name:            fixUpFuncName(funcName, awsServiceUpper),
		AWSName:         function.Name.Name,
		RecvType:        fmt.Sprintf("%[1]siface.%[2]sAPI", awsService, awsServiceUpper),
		ParamType:       g.expandTypeField(function.Type.Params),  // Assumes there is a single input parameter
		ResultType:      g.expandTypeField(function.Type.Results), // Assumes we can take the first return parameter
		InputPaginator:  g.inputPaginator,
		OutputPaginator: g.outputPaginator,
	}

	err := g.tmpl.Execute(&g.buf, funcSpec)
	if err != nil {
		log.Fatalf("error writing function \"%s\": %s", functionName, err)
	}
}

func (g *Generator) expandTypeField(field *ast.FieldList) string {
	typeValue := field.List[0].Type
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

func fixUpFuncName(funcName, service string) string {
	return strings.ReplaceAll(fixSomeInitialisms(funcName), service, "")
}

//go:embed header.tmpl
var headerTemplate string

//go:embed function.tmpl
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
	replace := s

	replace = strings.Replace(replace, "ResourceSes", "ResourceSES", 1)
	replace = strings.Replace(replace, "ApiGateway", "APIGateway", 1)
	replace = strings.Replace(replace, "Cloudwatch", "CloudWatch", 1)
	replace = strings.Replace(replace, "CurReport", "CURReport", 1)
	replace = strings.Replace(replace, "CloudHsm", "CloudHSM", 1)
	replace = strings.Replace(replace, "DynamoDb", "DynamoDB", 1)
	replace = strings.Replace(replace, "Opsworks", "OpsWorks", 1)
	replace = strings.Replace(replace, "Precheck", "PreCheck", 1)
	replace = strings.Replace(replace, "Graphql", "GraphQL", 1)
	replace = strings.Replace(replace, "Haproxy", "HAProxy", 1)
	replace = strings.Replace(replace, "Acmpca", "ACMPCA", 1)
	replace = strings.Replace(replace, "AcmPca", "ACMPCA", 1)
	replace = strings.Replace(replace, "Dnssec", "DNSSEC", 1)
	replace = strings.Replace(replace, "DocDb", "DocDB", 1)
	replace = strings.Replace(replace, "Docdb", "DocDB", 1)
	replace = strings.Replace(replace, "Https", "HTTPS", 1)
	replace = strings.Replace(replace, "Ipset", "IPSet", 1)
	replace = strings.Replace(replace, "Iscsi", "iSCSI", 1)
	replace = strings.Replace(replace, "Mysql", "MySQL", 1)
	replace = strings.Replace(replace, "Wafv2", "WAFV2", 1)
	replace = strings.Replace(replace, "Cidr", "CIDR", 1)
	replace = strings.Replace(replace, "Coip", "CoIP", 1)
	replace = strings.Replace(replace, "Dhcp", "DHCP", 1)
	replace = strings.Replace(replace, "Dkim", "DKIM", 1)
	replace = strings.Replace(replace, "Grpc", "GRPC", 1)
	replace = strings.Replace(replace, "Http", "HTTP", 1)
	replace = strings.Replace(replace, "Mwaa", "MWAA", 1)
	replace = strings.Replace(replace, "Oidc", "OIDC", 1)
	replace = strings.Replace(replace, "Qldb", "QLDB", 1)
	replace = strings.Replace(replace, "Smtp", "SMTP", 1)
	replace = strings.Replace(replace, "Xray", "XRay", 1)
	replace = strings.Replace(replace, "Acl", "ACL", 1)
	replace = strings.Replace(replace, "Acm", "ACM", 1)
	replace = strings.Replace(replace, "Ami", "AMI", 1)
	replace = strings.Replace(replace, "Api", "API", 1)
	replace = strings.Replace(replace, "Arn", "ARN", 1)
	replace = strings.Replace(replace, "Bgp", "BGP", 1)
	replace = strings.Replace(replace, "Csv", "CSV", 1)
	replace = strings.Replace(replace, "Dax", "DAX", 1)
	replace = strings.Replace(replace, "Dlm", "DLM", 1)
	replace = strings.Replace(replace, "Dms", "DMS", 1)
	replace = strings.Replace(replace, "Dns", "DNS", 1)
	replace = strings.Replace(replace, "Ebs", "EBS", 1)
	replace = strings.Replace(replace, "Ec2", "EC2", 1)
	replace = strings.Replace(replace, "Ecr", "ECR", 1)
	replace = strings.Replace(replace, "Ecs", "ECS", 1)
	replace = strings.Replace(replace, "Efs", "EFS", 1)
	replace = strings.Replace(replace, "Eip", "EIP", 1)
	replace = strings.Replace(replace, "Eks", "EKS", 1)
	replace = strings.Replace(replace, "Elb", "ELB", 1)
	replace = strings.Replace(replace, "Emr", "EMR", 1)
	replace = strings.Replace(replace, "Fms", "FMS", 1)
	replace = strings.Replace(replace, "Fsx", "FSx", 1)
	replace = strings.Replace(replace, "Hsm", "HSM", 1)
	replace = strings.Replace(replace, "Iam", "IAM", 1)
	replace = strings.Replace(replace, "Iot", "IoT", 1)
	replace = strings.Replace(replace, "Kms", "KMS", 1)
	replace = strings.Replace(replace, "Msk", "MSK", 1)
	replace = strings.Replace(replace, "Nat", "NAT", 1)
	replace = strings.Replace(replace, "Nfs", "NFS", 1)
	replace = strings.Replace(replace, "Php", "PHP", 1)
	replace = strings.Replace(replace, "Ram", "RAM", 1)
	replace = strings.Replace(replace, "Rds", "RDS", 1)
	replace = strings.Replace(replace, "Rfc", "RFC", 1)
	replace = strings.Replace(replace, "Sfn", "SFN", 1)
	replace = strings.Replace(replace, "Smb", "SMB", 1)
	replace = strings.Replace(replace, "Sms", "SMS", 1)
	replace = strings.Replace(replace, "Sns", "SNS", 1)
	replace = strings.Replace(replace, "Sql", "SQL", 1)
	replace = strings.Replace(replace, "Sqs", "SQS", 1)
	replace = strings.Replace(replace, "Ssh", "SSH", 1)
	replace = strings.Replace(replace, "Ssm", "SSM", 1)
	replace = strings.Replace(replace, "Sso", "SSO", 1)
	replace = strings.Replace(replace, "Sts", "STS", 1)
	replace = strings.Replace(replace, "Swf", "SWF", 1)
	replace = strings.Replace(replace, "Tcp", "TCP", 1)
	replace = strings.Replace(replace, "Vpc", "VPC", 1)
	replace = strings.Replace(replace, "Vpn", "VPN", 1)
	replace = strings.Replace(replace, "Waf", "WAF", 1)
	replace = strings.Replace(replace, "Xss", "XSS", 1)
	replace = strings.Replace(replace, "Db", "DB", 1)
	replace = strings.Replace(replace, "Ip", "IP", 1)
	replace = strings.Replace(replace, "Mq", "MQ", 1)

	if replace != strings.TrimSuffix(replace, "Ids") {
		replace = fmt.Sprintf("%s%s", strings.TrimSuffix(replace, "Ids"), "IDs")
	}

	if replace != strings.TrimSuffix(replace, "Id") {
		replace = fmt.Sprintf("%s%s", strings.TrimSuffix(replace, "Id"), "ID")
	}

	return replace
}
