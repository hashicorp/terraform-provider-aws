// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"maps"
	"path/filepath"
	"slices"
	"testing"
)

// testdataRoot is the root of a fixture shaped like the provider repository. See testdata/README.md.
const testdataRoot = "testdata"

func TestLoad(t *testing.T) {
	t.Parallel()

	a, err := load(testdataRoot)
	if err != nil {
		t.Fatalf("load(%q): %s", testdataRoot, err)
	}

	wantClientMethods := map[string]string{
		"ExampleBuildConstraintsClient": "github.com/aws/aws-sdk-go-v2/service/examplebuildconstraints",
		"ExampleCRUDClient":             "github.com/aws/aws-sdk-go-v2/service/examplecrud",
		"ExampleCrossServiceClient":     "github.com/aws/aws-sdk-go-v2/service/examplecrossservice",
		"ExampleCustomizeDiffClient":    "github.com/aws/aws-sdk-go-v2/service/examplecustomizediff",
		"ExampleFuncRefsClient":         "github.com/aws/aws-sdk-go-v2/service/examplefuncrefs",
		"ExampleImportClient":           "github.com/aws/aws-sdk-go-v2/service/exampleimport",
		"ExampleStructFieldClient":      "github.com/aws/aws-sdk-go-v2/service/examplestructfield",
		"ExampleTagsClient":             "github.com/aws/aws-sdk-go-v2/service/exampletags",
		"ExampleTagsGenClient":          "github.com/aws/aws-sdk-go-v2/service/exampletagsgen",
		"ExampleTagsUpdateOnlyClient":   "github.com/aws/aws-sdk-go-v2/service/exampletagsupdateonly",
	}
	if !maps.Equal(a.clientMethods, wantClientMethods) {
		t.Errorf("client methods = %v, want %v", a.clientMethods, wantClientMethods)
	}

	// Directories with no buildable Go files, or only test files, are not service packages.
	gotFiles := make(map[string][]string)
	for importPath, pkg := range a.services {
		gotFiles[importPath] = fileNames(a, pkg.files)
	}
	wantFiles := map[string][]string{
		servicePathPrefix + "examplebuildconstraints": {
			"included.go",
		},
		servicePathPrefix + "examplecrossservice": {
			"exports.go",
			"thing.go",
		},
		servicePathPrefix + "examplecrud": {
			"find.go",
			"gadget.go",
			"gizmo.go",
			"status.go",
			"wait.go",
			"widget.go",
		},
		servicePathPrefix + "examplecustomizediff": {
			"customize_diff.go",
			"thing.go",
		},
		servicePathPrefix + "examplefuncrefs": {
			"call_expression.go",
			"method_value.go",
		},
		servicePathPrefix + "exampleimport": {
			"import.go",
			"thing.go",
		},
		servicePathPrefix + "examplestructfield": {
			"database.go",
			"orchestrator.go",
			"parent_lookup.go",
			"retryer.go",
			"rule.go",
		},
		servicePathPrefix + "exampletags": {
			"gadget.go",
			"service_package.go",
			"tags.go",
			"widget.go",
			"without_identifier.go",
		},
		servicePathPrefix + "exampletagsgen": {
			"exports.go",
			"service_package.go",
			"tags_gen.go",
			"thing.go",
		},
		servicePathPrefix + "exampletagsupdateonly": {
			"service_package.go",
			"tags_gen.go",
			"thing.go",
		},
	}
	if !maps.EqualFunc(gotFiles, wantFiles, slices.Equal) {
		t.Errorf("service package files = %v, want %v", gotFiles, wantFiles)
	}

	for importPath, pkg := range a.services {
		if want := filepath.Base(importPath); pkg.name != want {
			t.Errorf("service package %s name = %q, want %q", importPath, pkg.name, want)
		}
	}
}

func TestLoadInvalidRoot(t *testing.T) {
	t.Parallel()

	root := filepath.Join(testdataRoot, "nonexistent")

	if _, err := load(root); err == nil {
		t.Errorf("load(%q) succeeded, want error", root)
	}
}

func TestParsePackage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		dir       string
		wantFiles []string
		wantErr   bool
	}{
		"conns": {
			dir:       "internal/conns",
			wantFiles: []string{"awsclient.go"},
		},
		"excludes test files and unsatisfied build constraints": {
			dir:       "internal/service/examplebuildconstraints",
			wantFiles: []string{"included.go"},
		},
		"several files": {
			dir: "internal/service/examplecrud",
			wantFiles: []string{
				"find.go",
				"gadget.go",
				"gizmo.go",
				"status.go",
				"wait.go",
				"widget.go",
			},
		},
		"no buildable Go files": {
			dir: "internal/service/examplenogofiles",
		},
		"only test files": {
			dir: "internal/service/exampletestonly",
		},
		"nonexistent directory": {
			dir:     "internal/service/examplenonexistent",
			wantErr: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			a := newTestAnalyzer()
			dir := filepath.Join(testdataRoot, filepath.FromSlash(testCase.dir))

			files, err := a.parsePackage(dir)

			if testCase.wantErr {
				if err == nil {
					t.Errorf("parsePackage(%q) succeeded, want error", dir)
				}
				return
			}
			if err != nil {
				t.Fatalf("parsePackage(%q): %s", dir, err)
			}
			if got := fileNames(a, files); !slices.Equal(got, testCase.wantFiles) {
				t.Errorf("parsePackage(%q) files = %q, want %q", dir, got, testCase.wantFiles)
			}
		})
	}
}

func TestParsePackageImports(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		dir  string
		file string
		want map[string]string
	}{
		"aliases and default names": {
			dir:  "internal/service/examplecrud",
			file: "gadget.go",
			want: map[string]string{
				"aws":                   "github.com/aws/aws-sdk-go-v2/aws",
				"awstypes":              "github.com/aws/aws-sdk-go-v2/service/examplecrud/types",
				"conns":                 "github.com/hashicorp/terraform-provider-aws/internal/conns",
				"context":               "context",
				"diag":                  "github.com/hashicorp/terraform-plugin-sdk/v2/diag",
				"errs":                  "github.com/hashicorp/terraform-provider-aws/internal/errs",
				"examplecrud":           "github.com/aws/aws-sdk-go-v2/service/examplecrud",
				"log":                   "log",
				"retry":                 "github.com/hashicorp/terraform-provider-aws/internal/retry",
				"schema":                "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema",
				"sdkdiag":               "github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag",
				"tfexamplecrossservice": "github.com/hashicorp/terraform-provider-aws/internal/service/examplecrossservice",
				"tfresource":            "github.com/hashicorp/terraform-provider-aws/internal/tfresource",
				"time":                  "time",
			},
		},
		"imports are recorded per file": {
			dir:  "internal/service/examplecrud",
			file: "status.go",
			want: map[string]string{
				"context":     "context",
				"examplecrud": "github.com/aws/aws-sdk-go-v2/service/examplecrud",
				"retry":       "github.com/hashicorp/terraform-provider-aws/internal/retry",
			},
		},
		"file with build constraint": {
			dir:  "internal/service/examplebuildconstraints",
			file: "included.go",
			want: map[string]string{
				"aws":                     "github.com/aws/aws-sdk-go-v2/aws",
				"awstypes":                "github.com/aws/aws-sdk-go-v2/service/examplebuildconstraints/types",
				"conns":                   "github.com/hashicorp/terraform-provider-aws/internal/conns",
				"context":                 "context",
				"diag":                    "github.com/hashicorp/terraform-plugin-sdk/v2/diag",
				"errs":                    "github.com/hashicorp/terraform-provider-aws/internal/errs",
				"examplebuildconstraints": "github.com/aws/aws-sdk-go-v2/service/examplebuildconstraints",
				"log":                     "log",
				"schema":                  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema",
				"sdkdiag":                 "github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag",
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			a := newTestAnalyzer()
			dir := filepath.Join(testdataRoot, filepath.FromSlash(testCase.dir))

			files, err := a.parsePackage(dir)
			if err != nil {
				t.Fatalf("parsePackage(%q): %s", dir, err)
			}

			i := slices.IndexFunc(fileNames(a, files), func(name string) bool { return name == testCase.file })
			if i < 0 {
				t.Fatalf("parsePackage(%q) did not parse %s", dir, testCase.file)
			}

			if got := a.importsOf(files[i]); !maps.Equal(got, testCase.want) {
				t.Errorf("imports of %s = %v, want %v", testCase.file, got, testCase.want)
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
		doc           []string // nil means the function has no doc comment.
		wantTypeNames []string
		wantTagged    bool
	}{
		"no doc": {},
		"no annotations": {
			doc: []string{
				"// resourceBucket returns the aws_s3_bucket resource.",
			},
		},
		"unrelated annotations": {
			doc: []string{
				`// @FrameworkResource("aws_s3_bucket", name="Bucket")`,
				"// @Region(global=true)",
				`// @ArnIdentity(identityDuplicateAttributes="id")`,
			},
		},
		"quoted type name": {
			doc: []string{
				`// @SDKResource("aws_s3_bucket", name="Bucket")`,
			},
			wantTypeNames: []string{"aws_s3_bucket"},
		},
		"unquoted type name": {
			doc: []string{
				"// @SDKResource(aws_s3_bucket, name=Bucket)",
			},
			wantTypeNames: []string{"aws_s3_bucket"},
		},
		"named argument before type name": {
			doc: []string{
				`// @SDKResource(name="Bucket", "aws_s3_bucket")`,
			},
			wantTypeNames: []string{"aws_s3_bucket"},
		},
		"named arguments around type name": {
			doc: []string{
				`// @SDKResource(name="Bucket", "aws_s3_bucket", tags=true)`,
			},
			wantTypeNames: []string{"aws_s3_bucket"},
		},
		"no space after comment marker": {
			doc: []string{
				`//@SDKResource("aws_s3_bucket", name="Bucket")`,
			},
			wantTypeNames: []string{"aws_s3_bucket"},
		},
		"trailing whitespace": {
			doc: []string{
				"// @SDKResource(\"aws_s3_bucket\", name=\"Bucket\")  \t",
			},
			wantTypeNames: []string{"aws_s3_bucket"},
		},
		"text after annotation": {
			doc: []string{
				`// @SDKResource("aws_s3_bucket", name="Bucket") is the bucket resource.`,
			},
		},
		"annotation without arguments": {
			doc: []string{
				"// @SDKResource",
			},
		},
		"annotation with empty arguments": {
			doc: []string{
				"// @SDKResource()",
			},
		},
		"annotation with only named arguments": {
			doc: []string{
				`// @SDKResource(name="Bucket")`,
			},
		},
		"type name without aws prefix": {
			doc: []string{
				`// @SDKResource("s3_bucket", name="Bucket")`,
			},
		},
		"type name with uppercase letters": {
			doc: []string{
				`// @SDKResource("aws_S3_bucket", name="Bucket")`,
			},
		},
		"type name with trailing underscore": {
			doc: []string{
				`// @SDKResource("aws_s3_bucket_", name="Bucket")`,
			},
		},
		"invalid first positional argument": {
			doc: []string{
				`// @SDKResource("s3_bucket", "aws_s3_bucket", name="Bucket")`,
			},
		},
		"multiple type names": {
			doc: []string{
				`// @SDKResource("aws_alb", name="Load Balancer")`,
				`// @SDKResource("aws_lb", name="Load Balancer")`,
			},
			wantTypeNames: []string{"aws_alb", "aws_lb"},
		},
		"multiple annotations with one invalid type name": {
			doc: []string{
				`// @SDKResource("aws_alb", name="Load Balancer")`,
				`// @SDKResource("lb", name="Load Balancer")`,
			},
			wantTypeNames: []string{"aws_alb"},
		},
		"annotations among other comments": {
			doc: []string{
				"// resourceLoadBalancer returns the load balancer resource.",
				"//",
				`// @SDKResource("aws_alb", name="Load Balancer")`,
				"// @Region(global=false)",
				`// @SDKResource("aws_lb", name="Load Balancer")`,
			},
			wantTypeNames: []string{"aws_alb", "aws_lb"},
		},
		"tags with identifier attribute": {
			doc: []string{
				`// @SDKResource("aws_s3_bucket", name="Bucket")`,
				`// @Tags(identifierAttribute="bucket")`,
			},
			wantTypeNames: []string{"aws_s3_bucket"},
			wantTagged:    true,
		},
		"tags with identifier attribute and resource type": {
			doc: []string{
				`// @SDKResource("aws_s3_object", name="Object")`,
				`// @Tags(identifierAttribute="arn", resourceType="Object")`,
			},
			wantTypeNames: []string{"aws_s3_object"},
			wantTagged:    true,
		},
		"tags without arguments": {
			doc: []string{
				`// @SDKResource("aws_codebuild_project", name="Project")`,
				"// @Tags",
			},
			wantTypeNames: []string{"aws_codebuild_project"},
		},
		"tags with only resource type": {
			doc: []string{
				`// @SDKResource("aws_s3_object", name="Object")`,
				`// @Tags(resourceType="Object")`,
			},
			wantTypeNames: []string{"aws_s3_object"},
		},
		"tags with identifier attribute before tags without": {
			doc: []string{
				`// @SDKResource("aws_s3_bucket", name="Bucket")`,
				`// @Tags(identifierAttribute="bucket")`,
				"// @Tags",
			},
			wantTypeNames: []string{"aws_s3_bucket"},
			wantTagged:    true,
		},
		"tags without type name": {
			doc: []string{
				`// @Tags(identifierAttribute="arn")`,
			},
			wantTagged: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gotTypeNames, gotTagged := parseAnnotations(toCommentGroup(testCase.doc))

			if !slices.Equal(gotTypeNames, testCase.wantTypeNames) {
				t.Errorf("type names = %q, want %q", gotTypeNames, testCase.wantTypeNames)
			}
			if gotTagged != testCase.wantTagged {
				t.Errorf("tagged = %t, want %t", gotTagged, testCase.wantTagged)
			}
		})
	}
}

func TestIsCustomDiffCombinator(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expr string
		want bool
	}{
		"all": {
			expr: "customdiff.All(resourceBucketCustomizeDiff, verify.SetTagsDiff)",
			want: true,
		},
		"sequence": {
			expr: "customdiff.Sequence(resourceBucketCustomizeDiff, verify.SetTagsDiff)",
			want: true,
		},
		"all without arguments": {
			expr: "customdiff.All()",
			want: true,
		},
		"nested": {
			expr: "customdiff.All(customdiff.Sequence(resourceBucketCustomizeDiff))",
			want: true,
		},
		"other customdiff function": {
			expr: `customdiff.ForceNewIfChange("size", sizeDecreased)`,
		},
		"conditional": {
			expr: "customdiff.If(isNewResource, resourceBucketCustomizeDiff)",
		},
		"combinator not called": {
			expr: "customdiff.All",
		},
		"different package": {
			expr: "diff.All(resourceBucketCustomizeDiff, verify.SetTagsDiff)",
		},
		"unqualified call": {
			expr: "All(resourceBucketCustomizeDiff, verify.SetTagsDiff)",
		},
		"ordinary call": {
			expr: "resourceBucketCustomizeDiff()",
		},
		"identifier": {
			expr: "resourceBucketCustomizeDiff",
		},
		"selector": {
			expr: "verify.SetTagsDiff",
		},
		"function literal": {
			expr: "func(ctx context.Context, d *schema.ResourceDiff, meta any) error { return nil }",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := isCustomDiffCombinator(mustParseExpr(t, testCase.expr)); got != testCase.want {
				t.Errorf("isCustomDiffCombinator(%s) = %t, want %t", testCase.expr, got, testCase.want)
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
		"s3c":      "github.com/aws/aws-sdk-go-v2/service/s3control",
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
		"pointer to pointer to SDK client": {
			expr: "**s3.Client",
			want: "github.com/aws/aws-sdk-go-v2/service/s3",
		},
		"SDK client with import alias": {
			expr: "*s3c.Client",
			want: "github.com/aws/aws-sdk-go-v2/service/s3control",
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
		"predeclared type": {
			expr: "string",
			want: "string",
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
		"map of SDK clients": {
			expr: "map[string]*s3.Client",
		},
		"function returning SDK client": {
			expr: "func() *s3.Client",
		},
		"interface literal": {
			expr: "interface{ ListTags() }",
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
		"service with digits": {
			importPath: "github.com/aws/aws-sdk-go-v2/service/ec2",
			want:       true,
		},
		"empty": {},
		"service directory": {
			importPath: "github.com/aws/aws-sdk-go-v2/service",
		},
		"aws": {
			importPath: "github.com/aws/aws-sdk-go-v2/aws",
		},
		"aws subpackage": {
			importPath: "github.com/aws/aws-sdk-go-v2/aws/arn",
		},
		"feature": {
			importPath: "github.com/aws/aws-sdk-go-v2/feature/s3/manager",
		},
		"smithy": {
			importPath: "github.com/aws/smithy-go",
		},
		"provider service package": {
			importPath: "github.com/hashicorp/terraform-provider-aws/internal/service/s3",
		},
		"package name": {
			importPath: "s3",
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

// fileNames returns the base names of parsed files.
func fileNames(a *analyzer, files []*ast.File) []string {
	var names []string
	for _, file := range files {
		names = append(names, filepath.Base(a.fset.File(file.Pos()).Name()))
	}
	return names
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
