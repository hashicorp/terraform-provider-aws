// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"go/ast"
	"go/parser"
	"go/types"
	"slices"
	"testing"
)

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

// mustParseExpr parses a Go expression, failing the test if it is invalid.
func mustParseExpr(t *testing.T, src string) ast.Expr {
	t.Helper()

	expr, err := parser.ParseExpr(src)
	if err != nil {
		t.Fatalf("parsing %q: %s", src, err)
	}
	return expr
}
