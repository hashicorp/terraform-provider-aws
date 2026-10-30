// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"encoding/json"
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// testdataRoot is the root of a fixture shaped like the provider repository. See testdata/README.md.
const testdataRoot = "testdata"

// goldenDir holds the golden files that TestMapPackages compares results with.
var goldenDir = filepath.Join(testdataRoot, "golden")

var update = flag.Bool("update", false, "update the golden files in "+goldenDir)

func TestLoadInvalidRoot(t *testing.T) {
	t.Parallel()

	root := filepath.Join(testdataRoot, "nonexistent")

	if _, err := load(root); err == nil {
		t.Errorf("load(%q) succeeded, want error", root)
	}
}

func TestParsePackageInvalidDir(t *testing.T) {
	t.Parallel()

	a := newTestAnalyzer()
	dir := filepath.Join(testdataRoot, "internal", "service", "mocknonexistent")

	if _, err := a.parsePackage(dir); err == nil {
		t.Errorf("parsePackage(%q) succeeded, want error", dir)
	}
}

func TestMapPackages(t *testing.T) {
	t.Parallel()

	a, err := load(testdataRoot)
	if err != nil {
		t.Fatalf("load(%q): %s", testdataRoot, err)
	}

	// The resources and data sources found in each service package are compared with the golden
	// file testdata/golden/<pkg>.json. See testdata/README.md.
	testCases := map[string]struct {
		pkg string
	}{
		"build constraints": {
			pkg: "mockbuildconstraints",
		},
		"package exporting a finder": {
			pkg: "mockcrossservice",
		},
		"CRUD": {
			pkg: "mockcrud",
		},
		"CustomizeDiff": {
			pkg: "mockcustomizediff",
		},
		"data sources": {
			pkg: "mockdatasource",
		},
		"function references": {
			pkg: "mockfuncrefs",
		},
		"importers": {
			pkg: "mockimport",
		},
		"interface dispatch": {
			pkg: "mockinterface",
		},
		"struct fields": {
			pkg: "mockstructfield",
		},
		"shared tag handlers branch on resource type": {
			pkg: "mocktags",
		},
		"generated tags": {
			pkg: "mocktagsgen",
		},
		"generated UpdateTags only": {
			pkg: "mocktagsupdateonly",
		},
		"tagged without ListTags or UpdateTags": {
			pkg: "mocktagsnohandlers",
		},
		"unanalyzed functions": {
			pkg: "mockunanalyzed",
		},
	}

	// Every service package, and every golden file, must have a test case.
	tested := make(map[string]bool)
	for _, testCase := range testCases {
		tested[testCase.pkg] = true
	}
	for _, pkg := range a.servicePackages() {
		if !tested[pkg.name] {
			t.Errorf("no test case for service package %s", pkg.name)
		}
	}
	goldenFiles, err := filepath.Glob(filepath.Join(goldenDir, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, goldenFile := range goldenFiles {
		if !tested[strings.TrimSuffix(filepath.Base(goldenFile), ".json")] {
			t.Errorf("no test case for golden file %s", goldenFile)
		}
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pkg, ok := a.services[servicePathPrefix+testCase.pkg]
			if !ok {
				t.Fatalf("service package %s not loaded", testCase.pkg)
			}

			m := a.mapPackages([]*servicePackage{pkg})

			// Order isn't asserted.
			slices.SortFunc(m.Resources, func(x, y resource) int {
				return strings.Compare(x.ResourceType, y.ResourceType)
			})
			slices.SortFunc(m.DataSources, func(x, y dataSource) int {
				return strings.Compare(x.DataSourceType, y.DataSourceType)
			})
			got, err := json.MarshalIndent(m, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			got = append(got, '\n')

			goldenFile := filepath.Join(goldenDir, testCase.pkg+".json")
			if *update {
				if err := os.WriteFile(goldenFile, got, 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}

			want, err := os.ReadFile(goldenFile)
			if err != nil {
				t.Fatalf("reading golden file: %s; to create it, run go test -run TestMapPackages -update", err)
			}
			if line, gotLine, wantLine, ok := firstDifference(got, want); ok {
				t.Errorf("mapPackages(%s) differs from %s at line %d:\n got: %s\nwant: %s\nIf the change is intended, run go test -run TestMapPackages -update", testCase.pkg, goldenFile, line, gotLine, wantLine)
			}
		})
	}
}

func TestParseAnnotations(t *testing.T) {
	t.Parallel()

	toCommentGroup := func(doc []string) *ast.CommentGroup {
		if doc == nil {
			return nil
		}
		cg := &ast.CommentGroup{}
		for _, text := range doc {
			cg.List = append(cg.List, &ast.Comment{Text: text})
		}
		return cg
	}

	testCases := map[string]struct {
		doc                  []string // nil means the function has no doc comment.
		annotation           string
		wantTypeNames        []string
		wantTagged           bool
		wantTagsResourceType string
	}{
		"no doc": {
			annotation: annotationSDKResource,
		},
		"no annotations": {
			doc: []string{
				"// resourceBucket returns the aws_s3_bucket resource.",
			},
			annotation: annotationSDKResource,
		},
		"unrelated annotations": {
			doc: []string{
				`// @FrameworkResource("aws_s3_bucket", name="Bucket")`,
				"// @Region(global=true)",
				`// @ArnIdentity(identityDuplicateAttributes="id")`,
			},
			annotation: annotationSDKResource,
		},
		"quoted type name": {
			doc: []string{
				`// @SDKResource("aws_s3_bucket", name="Bucket")`,
			},
			annotation:    annotationSDKResource,
			wantTypeNames: []string{"aws_s3_bucket"},
		},
		"unquoted type name": {
			doc: []string{
				"// @SDKResource(aws_s3_bucket, name=Bucket)",
			},
			annotation:    annotationSDKResource,
			wantTypeNames: []string{"aws_s3_bucket"},
		},
		"named argument before type name": {
			doc: []string{
				`// @SDKResource(name="Bucket", "aws_s3_bucket")`,
			},
			annotation:    annotationSDKResource,
			wantTypeNames: []string{"aws_s3_bucket"},
		},
		"no space after comment marker": {
			doc: []string{
				`//@SDKResource("aws_s3_bucket", name="Bucket")`,
			},
			annotation:    annotationSDKResource,
			wantTypeNames: []string{"aws_s3_bucket"},
		},
		"trailing whitespace": {
			doc: []string{
				"// @SDKResource(\"aws_s3_bucket\", name=\"Bucket\")  \t",
			},
			annotation:    annotationSDKResource,
			wantTypeNames: []string{"aws_s3_bucket"},
		},
		"text after annotation": {
			doc: []string{
				`// @SDKResource("aws_s3_bucket", name="Bucket") is the bucket resource.`,
			},
			annotation: annotationSDKResource,
		},
		"annotation without arguments": {
			doc: []string{
				"// @SDKResource",
			},
			annotation: annotationSDKResource,
		},
		"annotation with empty arguments": {
			doc: []string{
				"// @SDKResource()",
			},
			annotation: annotationSDKResource,
		},
		"annotation with only named arguments": {
			doc: []string{
				`// @SDKResource(name="Bucket")`,
			},
			annotation: annotationSDKResource,
		},
		"type name without aws prefix": {
			doc: []string{
				`// @SDKResource("s3_bucket", name="Bucket")`,
			},
			annotation: annotationSDKResource,
		},
		"type name with uppercase letters": {
			doc: []string{
				`// @SDKResource("aws_S3_bucket", name="Bucket")`,
			},
			annotation: annotationSDKResource,
		},
		"type name with trailing underscore": {
			doc: []string{
				`// @SDKResource("aws_s3_bucket_", name="Bucket")`,
			},
			annotation: annotationSDKResource,
		},
		"invalid first positional argument": {
			doc: []string{
				`// @SDKResource("s3_bucket", "aws_s3_bucket", name="Bucket")`,
			},
			annotation: annotationSDKResource,
		},
		"annotations among other comments": {
			doc: []string{
				"// resourceLoadBalancer returns the load balancer resource.",
				"//",
				`// @SDKResource("aws_alb", name="Load Balancer")`,
				"// @Region(global=false)",
				`// @SDKResource("aws_lb", name="Load Balancer")`,
			},
			annotation:    annotationSDKResource,
			wantTypeNames: []string{"aws_alb", "aws_lb"},
		},
		"tags with identifier attribute and resource type": {
			doc: []string{
				`// @SDKResource("aws_s3_object", name="Object")`,
				`// @Tags(identifierAttribute="arn", resourceType="Object")`,
			},
			annotation:           annotationSDKResource,
			wantTypeNames:        []string{"aws_s3_object"},
			wantTagged:           true,
			wantTagsResourceType: "Object",
		},
		"tags with resource type before identifier attribute": {
			doc: []string{
				`// @SDKResource("aws_s3_object", name="Object")`,
				`// @Tags(resourceType="Object", identifierAttribute="arn")`,
			},
			annotation:           annotationSDKResource,
			wantTypeNames:        []string{"aws_s3_object"},
			wantTagged:           true,
			wantTagsResourceType: "Object",
		},
		"tags without arguments": {
			doc: []string{
				`// @SDKResource("aws_codebuild_project", name="Project")`,
				"// @Tags",
			},
			annotation:    annotationSDKResource,
			wantTypeNames: []string{"aws_codebuild_project"},
		},
		"tags with only resource type": {
			doc: []string{
				`// @SDKResource("aws_s3_object", name="Object")`,
				`// @Tags(resourceType="Object")`,
			},
			annotation:           annotationSDKResource,
			wantTypeNames:        []string{"aws_s3_object"},
			wantTagsResourceType: "Object",
		},
		"later tags resource type replaces earlier": {
			doc: []string{
				`// @SDKResource("aws_s3_object", name="Object")`,
				`// @Tags(identifierAttribute="arn", resourceType="Object")`,
				`// @Tags(resourceType="ObjectCopy")`,
			},
			annotation:           annotationSDKResource,
			wantTypeNames:        []string{"aws_s3_object"},
			wantTagged:           true,
			wantTagsResourceType: "ObjectCopy",
		},
		"tags with identifier attribute before tags without": {
			doc: []string{
				`// @SDKResource("aws_s3_bucket", name="Bucket")`,
				`// @Tags(identifierAttribute="bucket")`,
				"// @Tags",
			},
			annotation:    annotationSDKResource,
			wantTypeNames: []string{"aws_s3_bucket"},
			wantTagged:    true,
		},
		"tags without type name": {
			doc: []string{
				`// @Tags(identifierAttribute="arn")`,
			},
			annotation: annotationSDKResource,
			wantTagged: true,
		},
		"data source": {
			doc: []string{
				`// @SDKDataSource("aws_s3_bucket", name="Bucket")`,
			},
			annotation:    annotationSDKDataSource,
			wantTypeNames: []string{"aws_s3_bucket"},
		},
		"data source with several type names": {
			doc: []string{
				`// @SDKDataSource("aws_alb", name="Load Balancer")`,
				`// @SDKDataSource("aws_lb", name="Load Balancer")`,
			},
			annotation:    annotationSDKDataSource,
			wantTypeNames: []string{"aws_alb", "aws_lb"},
		},
		"data source with tags": {
			doc: []string{
				`// @SDKDataSource("aws_s3_object", name="Object")`,
				`// @Tags(identifierAttribute="arn", resourceType="Object")`,
			},
			annotation:           annotationSDKDataSource,
			wantTypeNames:        []string{"aws_s3_object"},
			wantTagged:           true,
			wantTagsResourceType: "Object",
		},
		"data source with invalid type name": {
			doc: []string{
				`// @SDKDataSource("s3_bucket", name="Bucket")`,
			},
			annotation: annotationSDKDataSource,
		},
		"data source annotation for resources": {
			doc: []string{
				`// @SDKDataSource("aws_s3_bucket", name="Bucket")`,
			},
			annotation: annotationSDKResource,
		},
		"resource annotation for data sources": {
			doc: []string{
				`// @SDKResource("aws_s3_bucket", name="Bucket")`,
				`// @Tags(identifierAttribute="bucket")`,
			},
			annotation: annotationSDKDataSource,
			wantTagged: true,
		},
		"framework data source annotation for data sources": {
			doc: []string{
				`// @FrameworkDataSource("aws_s3_bucket", name="Bucket")`,
			},
			annotation: annotationSDKDataSource,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gotTypeNames, gotTagged, gotTagsResourceType := parseAnnotations(toCommentGroup(testCase.doc), testCase.annotation)

			if !slices.Equal(gotTypeNames, testCase.wantTypeNames) {
				t.Errorf("type names = %q, want %q", gotTypeNames, testCase.wantTypeNames)
			}
			if gotTagged != testCase.wantTagged {
				t.Errorf("tagged = %t, want %t", gotTagged, testCase.wantTagged)
			}
			if gotTagsResourceType != testCase.wantTagsResourceType {
				t.Errorf("tags resource type = %q, want %q", gotTagsResourceType, testCase.wantTagsResourceType)
			}
		})
	}
}

func TestCustomizeDiffFuncs(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expr string
		// want holds the returned expressions, formatted by types.ExprString.
		want []string
	}{
		"identifier": {
			expr: "resourceBucketCustomizeDiff",
			want: []string{"resourceBucketCustomizeDiff"},
		},
		"selector": {
			expr: "verify.SetTagsDiff",
			want: []string{"verify.SetTagsDiff"},
		},
		"function literal": {
			expr: "func(ctx context.Context, d *schema.ResourceDiff, meta any) error { return nil }",
			want: []string{"(func(ctx context.Context, d *schema.ResourceDiff, meta any) error literal)"},
		},
		"ordinary call": {
			expr: "resourceBucketCustomizeDiff()",
			want: []string{"resourceBucketCustomizeDiff()"},
		},
		"other customdiff function": {
			expr: `customdiff.ForceNewIfChange("size", sizeDecreased)`,
			want: []string{`customdiff.ForceNewIfChange("size", sizeDecreased)`},
		},
		"combinator not called": {
			expr: "customdiff.All",
			want: []string{"customdiff.All"},
		},
		"combinator from another package": {
			expr: "diff.All(resourceBucketCustomizeDiff, verify.SetTagsDiff)",
			want: []string{"diff.All(resourceBucketCustomizeDiff, verify.SetTagsDiff)"},
		},
		"unqualified combinator": {
			expr: "All(resourceBucketCustomizeDiff, verify.SetTagsDiff)",
			want: []string{"All(resourceBucketCustomizeDiff, verify.SetTagsDiff)"},
		},
		"all": {
			expr: "customdiff.All(resourceBucketCustomizeDiff, verify.SetTagsDiff)",
			want: []string{"resourceBucketCustomizeDiff", "verify.SetTagsDiff"},
		},
		"sequence": {
			expr: "customdiff.Sequence(resourceBucketCustomizeDiff, verify.SetTagsDiff)",
			want: []string{"resourceBucketCustomizeDiff", "verify.SetTagsDiff"},
		},
		"all without arguments": {
			expr: "customdiff.All()",
		},
		"all with calls": {
			expr: `customdiff.All(customdiff.ForceNewIfChange("size", sizeDecreased), validateBucket())`,
			want: []string{`customdiff.ForceNewIfChange("size", sizeDecreased)`, "validateBucket()"},
		},
		"all with conditional": {
			expr: "customdiff.All(resourceBucketCustomizeDiff, customdiff.If(isNewResource, validateBucket))",
			want: []string{"resourceBucketCustomizeDiff", "customdiff.If(isNewResource, validateBucket)"},
		},
		"nested": {
			expr: "customdiff.All(first, customdiff.Sequence(second, third), fourth)",
			want: []string{"first", "second", "third", "fourth"},
		},
		"deeply nested": {
			expr: "customdiff.Sequence(customdiff.All(customdiff.Sequence(first), second), customdiff.All(), third)",
			want: []string{"first", "second", "third"},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var got []string
			for _, expr := range customizeDiffFuncs(mustParseExpr(t, testCase.expr)) {
				got = append(got, types.ExprString(expr))
			}

			if !slices.Equal(got, testCase.want) {
				t.Errorf("customizeDiffFuncs(%s) = %q, want %q", testCase.expr, got, testCase.want)
			}
		})
	}
}

func TestResolveType(t *testing.T) {
	t.Parallel()

	imports := map[string]string{
		"aws":      "github.com/aws/aws-sdk-go-v2/aws",
		"awstypes": "github.com/aws/aws-sdk-go-v2/service/s3/types",
		"conns":    "github.com/hashicorp/terraform-provider-aws/internal/conns",
		"s3":       "github.com/aws/aws-sdk-go-v2/service/s3",
		"tfs3":     "github.com/hashicorp/terraform-provider-aws/internal/service/s3",
	}

	testCases := map[string]struct {
		expr string
		want string
	}{
		"SDK client": {
			expr: "s3.Client",
			want: "github.com/aws/aws-sdk-go-v2/service/s3",
		},
		"pointer to SDK client": {
			expr: "*s3.Client",
			want: "github.com/aws/aws-sdk-go-v2/service/s3",
		},
		"SDK client from unimported package": {
			expr: "*ec2.Client",
		},
		"other SDK service type": {
			expr: "*s3.Options",
		},
		"SDK model type": {
			expr: "*awstypes.Bucket",
		},
		"non-service SDK type": {
			expr: "aws.Config",
		},
		"provider service package client": {
			expr: "*tfs3.Client",
		},
		"qualified provider type": {
			expr: "*conns.AWSClient",
		},
		"unqualified type": {
			expr: "AWSClient",
			want: "AWSClient",
		},
		"pointer to unqualified type": {
			expr: "*servicePackage",
			want: "servicePackage",
		},
		"generic type with one type argument": {
			expr: "*retryer[*s3.Client]",
			want: "retryer",
		},
		"generic type with several type arguments": {
			expr: "cache[string, *s3.Client]",
			want: "cache",
		},
		"slice of SDK clients": {
			expr: "[]*s3.Client",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := resolveType(mustParseExpr(t, testCase.expr), imports); got != testCase.want {
				t.Errorf("resolveType(%s) = %q, want %q", testCase.expr, got, testCase.want)
			}
		})
	}
}

func TestIsSDKPackage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		importPath string
		want       bool
	}{
		"service": {
			importPath: "github.com/aws/aws-sdk-go-v2/service/s3",
			want:       true,
		},
		"service directory": {
			importPath: "github.com/aws/aws-sdk-go-v2/service",
		},
		"aws": {
			importPath: "github.com/aws/aws-sdk-go-v2/aws",
		},
		"feature": {
			importPath: "github.com/aws/aws-sdk-go-v2/feature/s3/manager",
		},
		"provider service package": {
			importPath: "github.com/hashicorp/terraform-provider-aws/internal/service/s3",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := isSDKPackage(testCase.importPath); got != testCase.want {
				t.Errorf("isSDKPackage(%q) = %t, want %t", testCase.importPath, got, testCase.want)
			}
		})
	}
}

func TestCaseClauseFor(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		src   string // A switch statement.
		value string
		// wantClause is the index of the clause in the switch, or -1 for none.
		wantClause int
		wantOK     bool
	}{
		"matching case": {
			src: `switch resourceType {
			case "Role":
				role()
			case "User":
				user()
			}`,
			value:      "User",
			wantClause: 1,
			wantOK:     true,
		},
		"case listing several values": {
			src: `switch resourceType {
			case "Bucket":
				bucket()
			case "Object", "ObjectCopy", "BucketObject":
				object()
			}`,
			value:      "ObjectCopy",
			wantClause: 1,
			wantOK:     true,
		},
		"matching case without statements": {
			src: `switch resourceType {
			case "Role":
			default:
				other()
			}`,
			value:      "Role",
			wantClause: 0,
			wantOK:     true,
		},
		"default when no case matches": {
			src: `switch resourceType {
			case "Role":
				role()
			default:
				other()
			}`,
			value:      "User",
			wantClause: 1,
			wantOK:     true,
		},
		"matching case after default": {
			src: `switch resourceType {
			default:
				other()
			case "Role":
				role()
			}`,
			value:      "Role",
			wantClause: 1,
			wantOK:     true,
		},
		"no matching case and no default": {
			src: `switch resourceType {
			case "Role":
				role()
			}`,
			value:      "User",
			wantClause: -1,
			wantOK:     true,
		},
		"case that isn't a literal": {
			src: `switch resourceType {
			case "Role":
				role()
			case userType:
				user()
			}`,
			value:      "Role",
			wantClause: -1,
		},
		"case that isn't a string": {
			src: `switch resourceType {
			case 1:
				one()
			}`,
			value:      "Role",
			wantClause: -1,
		},
		"matching case falls through": {
			src: `switch resourceType {
			case "Role":
				role()
				fallthrough
			case "User":
				user()
			}`,
			value:      "Role",
			wantClause: -1,
		},
		"default falls through": {
			src: `switch resourceType {
			default:
				other()
				fallthrough
			case "Role":
				role()
			}`,
			value:      "User",
			wantClause: -1,
		},
		"other case falls through": {
			src: `switch resourceType {
			case "Role":
				role()
				fallthrough
			case "User":
				user()
			}`,
			value:      "User",
			wantClause: 1,
			wantOK:     true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := mustParseSwitch(t, testCase.src)

			clause, ok := caseClauseFor(s, testCase.value)

			gotClause := -1
			if clause != nil {
				gotClause = slices.Index(s.Body.List, ast.Stmt(clause))
			}
			if gotClause != testCase.wantClause || ok != testCase.wantOK {
				t.Errorf("caseClauseFor(%q) = clause %d, %t, want clause %d, %t", testCase.value, gotClause, ok, testCase.wantClause, testCase.wantOK)
			}
		})
	}
}

// newTestAnalyzer returns an analyzer for testdataRoot with nothing loaded.
func newTestAnalyzer() *analyzer {
	return &analyzer{
		root:          testdataRoot,
		fset:          token.NewFileSet(),
		imports:       make(map[*token.File]map[string]string),
		clientMethods: make(map[string]string),
		services:      make(map[string]*servicePackage),
	}
}

// mustParseExpr parses a Go expression, failing the test if it is invalid.
func mustParseExpr(t *testing.T, src string) ast.Expr {
	t.Helper()

	expr, err := parser.ParseExpr(src)
	if err != nil {
		t.Fatalf("parsing %q: %s", src, err)
	}
	return expr
}

// mustParseSwitch parses a Go switch statement, failing the test if it is invalid.
func mustParseSwitch(t *testing.T, src string) *ast.SwitchStmt {
	t.Helper()

	file, err := parser.ParseFile(token.NewFileSet(), "", "package p\n\nfunc f() {\n"+src+"\n}\n", 0)
	if err != nil {
		t.Fatalf("parsing %q: %s", src, err)
	}
	body := file.Decls[0].(*ast.FuncDecl).Body
	s, ok := body.List[0].(*ast.SwitchStmt)
	if !ok || len(body.List) != 1 {
		t.Fatalf("%q is not a single switch statement", src)
	}
	return s
}

// firstDifference returns the first line, numbered from 1, at which got and want differ, and that
// line of each. ok is false if they are equal.
func firstDifference(got, want []byte) (line int, gotLine, wantLine string, ok bool) {
	gotLines, wantLines := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	lineAt := func(lines []string, i int) string {
		if i < len(lines) {
			return lines[i]
		}
		return "(end of file)"
	}

	for i := range max(len(gotLines), len(wantLines)) {
		if g, w := lineAt(gotLines, i), lineAt(wantLines, i); g != w {
			return i + 1, g, w, true
		}
	}
	return 0, "", "", false
}
