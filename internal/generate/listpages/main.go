// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

//go:build generate

package main

import (
	_ "embed"
	"flag"
	"fmt"
	"go/ast"
	"os"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names/data"
)

var (
	inputPaginator  = flag.String("InputPaginator", "", "name of the input pagination token field")
	listOps         = flag.String("ListOps", "", "ListOps")
	outputPaginator = flag.String("OutputPaginator", "", "name of the output pagination token field")
	paginator       = flag.String("Paginator", "NextToken", "name of the pagination token field")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\tmain.go [flags] [<generated-lister-file>]\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

type TemplateData struct {
	Parameters      string
	AWSService      string
	ServicePackage  string
	InputPaginator  string
	OutputPaginator string
	Name            string
	AWSName         string
	ParamType       string
	ResultType      string
}

func main() {
	flag.Usage = usage
	flag.Parse()

	g := common.NewGenerator()

	if (*inputPaginator != "" && *outputPaginator == "") || (*inputPaginator == "" && *outputPaginator != "") {
		g.Fatalf("both InputPaginator and OutputPaginator must be specified if one is")
	}

	if *inputPaginator == "" {
		*inputPaginator = *paginator
	}
	if *outputPaginator == "" {
		*outputPaginator = *paginator
	}

	filename := "list_pages_gen.go"
	if args := flag.Args(); len(args) > 0 {
		filename = args[0]
	}

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

	d := g.NewGoFileDestination(filename)

	templateData := TemplateData{
		Parameters:      strings.Join(os.Args[1:], " "),
		AWSService:      awsPkg,
		ServicePackage:  servicePackage,
		InputPaginator:  *inputPaginator,
		OutputPaginator: *outputPaginator,
	}

	if err := d.BufferTemplate("header", headerTemplate, templateData); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	functions := strings.Split(*listOps, ",")
	slices.Sort(functions)
	for _, functionName := range functions {
		function := pkg.FindFunction(functionName)
		if function == nil {
			g.Fatalf("function (%q) not found", functionName)
		}

		awsFunctionName := function.Name()
		localFunctionName := localFunctionName(awsFunctionName)

		g.Infof("  %s.%s", servicePackage, localFunctionName)

		templateData.Name = localFunctionName
		templateData.AWSName = awsFunctionName

		params, results := function.Params(), function.Results()
		// 1st parameter should be context.Context.
		// 2nd parameter should be *<Service>.<FunctionName>Input.
		if len(params) < 2 {
			g.Fatalf("%d params found", len(params))
		}
		expr := params[1].Type
		name := typeName(expr)
		if name == "" {
			g.Fatalf("unexpected type expression: (%[1]T) %[1]v", expr)
		}
		templateData.ParamType = name

		// 1st result should be *<Service>.<FunctionName>Output.
		// 2nd result should be error.
		if len(results) < 2 {
			g.Fatalf("%d results found", len(results))
		}
		expr = results[0].Type
		name = typeName(expr)
		if name == "" {
			g.Fatalf("unexpected type expression: (%[1]T) %[1]v", expr)
		}
		templateData.ResultType = name

		if err := d.BufferTemplate("function", functionTemplate, templateData); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

func identifierName(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

func typeName(expr ast.Expr) string {
	if star, ok := expr.(*ast.StarExpr); ok {
		return identifierName(star.X)
	}
	return ""
}

func localFunctionName(awsFunctionName string) string {
	return fixSomeInitialisms(strings.ToLower(awsFunctionName[0:1]) + awsFunctionName[1:])
}

//go:embed header.gtpl
var headerTemplate string

//go:embed function.gtpl
var functionTemplate string

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
