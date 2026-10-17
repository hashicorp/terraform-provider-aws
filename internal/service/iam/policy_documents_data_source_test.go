// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestSplitPolicyDocument(t *testing.T) {
	t.Parallel()

	statement := func(sid string) *tfiam.IAMPolicyStatement {
		return &tfiam.IAMPolicyStatement{
			Sid:       sid,
			Effect:    "Allow",
			Actions:   "s3:GetObject",
			Resources: "*",
		}
	}
	doc := func(sids ...string) *tfiam.IAMPolicyDoc {
		d := &tfiam.IAMPolicyDoc{Version: "2012-10-17", Id: "test"}
		for _, sid := range sids {
			d.Statements = append(d.Statements, statement(sid))
		}
		return d
	}
	size := func(d *tfiam.IAMPolicyDoc) int {
		b, err := json.Marshal(d)
		if err != nil {
			t.Fatal(err)
		}
		return len(b)
	}

	// All statements have the same length, so the size of an N-statement
	// document is the same whichever Sids it holds.
	oneSize := size(doc("S1"))
	twoSize := size(doc("S1", "S2"))

	testCases := map[string]struct {
		doc         *tfiam.IAMPolicyDoc
		maxSize     int
		expected    [][]string
		expectedErr string
	}{
		"empty": {
			doc:      doc(),
			maxSize:  6144,
			expected: [][]string{{}},
		},
		"empty too large": {
			doc:         doc(),
			maxSize:     1,
			expectedErr: "without statements",
		},
		"fits in one": {
			doc:      doc("S1", "S2", "S3"),
			maxSize:  6144,
			expected: [][]string{{"S1", "S2", "S3"}},
		},
		"exact boundary": {
			doc:      doc("S1", "S2"),
			maxSize:  twoSize,
			expected: [][]string{{"S1", "S2"}},
		},
		"one under boundary": {
			doc:      doc("S1", "S2"),
			maxSize:  twoSize - 1,
			expected: [][]string{{"S1"}, {"S2"}},
		},
		"splits preserving order": {
			doc:      doc("S1", "S2", "S3", "S4", "S5"),
			maxSize:  twoSize,
			expected: [][]string{{"S1", "S2"}, {"S3", "S4"}, {"S5"}},
		},
		"oversized first statement": {
			doc:         doc("S1", "S2"),
			maxSize:     oneSize - 1,
			expectedErr: `statement 0 (Sid "S1")`,
		},
		"oversized later statement": {
			doc: &tfiam.IAMPolicyDoc{
				Version: "2012-10-17",
				Statements: []*tfiam.IAMPolicyStatement{
					statement("S1"),
					{Effect: "Allow", Actions: "s3:GetObject", Resources: strings.Repeat("x", 1000)},
				},
			},
			maxSize:     oneSize,
			expectedErr: "statement 1 of the merged policy document",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := tfiam.SplitPolicyDocument(testCase.doc, testCase.maxSize)

			if testCase.expectedErr != "" {
				if err == nil || !strings.Contains(err.Error(), testCase.expectedErr) {
					t.Fatalf("expected error containing %q, got %v", testCase.expectedErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if len(got) != len(testCase.expected) {
				t.Fatalf("expected %d documents, got %d", len(testCase.expected), len(got))
			}
			for i, d := range got {
				if d.Version != testCase.doc.Version || d.Id != testCase.doc.Id {
					t.Errorf("document %d: expected Version %q and Id %q, got %q and %q", i, testCase.doc.Version, testCase.doc.Id, d.Version, d.Id)
				}
				if s := size(d); s > testCase.maxSize {
					t.Errorf("document %d: size %d exceeds %d", i, s, testCase.maxSize)
				}
				var sids []string
				for _, stmt := range d.Statements {
					sids = append(sids, stmt.Sid)
				}
				if got, want := strings.Join(sids, ","), strings.Join(testCase.expected[i], ","); got != want {
					t.Errorf("document %d: expected Sids %q, got %q", i, want, got)
				}
			}
		})
	}
}

func TestAccIAMPolicyDocumentsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_policy_documents.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentsDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "json.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "minified_json.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "minified_json.0", fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Sid":"S1","Effect":"Allow","Action":["s3:PutObject","s3:GetObject"],"Resource":"arn:%s:s3:::bucket1/*"}]}`, acctest.Partition())),
					acctest.CheckResourceAttrEquivalentJSON(dataSourceName, "json.0", fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Sid":"S1","Effect":"Allow","Action":["s3:PutObject","s3:GetObject"],"Resource":"arn:%s:s3:::bucket1/*"}]}`, acctest.Partition())),
					resource.TestCheckResourceAttr(dataSourceName, "max_policy_size", "6144"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrVersion, "2012-10-17"),
					resource.TestCheckResourceAttr(dataSourceName, "statement.0.effect", "Allow"),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentsDataSource_split(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_policy_documents.test"
	twoStatementSize := 218 + 2*(len(acctest.Partition())-len("aws"))

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				// A two-statement document exactly fits; one byte less requires a split.
				Config: testAccPolicyDocumentsDataSourceConfig_split(twoStatementSize),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "json.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "minified_json.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "minified_json.0", fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Sid":"S1","Effect":"Allow","Action":"s3:GetObject","Resource":"arn:%[1]s:s3:::bucket1/*"},{"Sid":"S2","Effect":"Allow","Action":"s3:GetObject","Resource":"arn:%[1]s:s3:::bucket2/*"}]}`, acctest.Partition())),
					resource.TestCheckResourceAttr(dataSourceName, "minified_json.1", fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Sid":"S3","Effect":"Allow","Action":"s3:GetObject","Resource":"arn:%s:s3:::bucket3/*"}]}`, acctest.Partition())),
				),
			},
			{
				Config: testAccPolicyDocumentsDataSourceConfig_split(twoStatementSize - 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "minified_json.#", "3"),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentsDataSource_sourceOverride(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_policy_documents.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentsDataSourceConfig_sourceOverride,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "minified_json.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "minified_json.0", `{"Version":"2012-10-17","Statement":[{"Sid":"Source","Effect":"Deny","Action":"s3:DeleteObject","Resource":"*"},{"Sid":"Config","Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}`),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentsDataSource_emptyString(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_policy_documents.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentsDataSourceConfig_emptyString,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "minified_json.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "minified_json.0", `{"Version":"2012-10-17"}`),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentsDataSource_duplicateSid(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPolicyDocumentsDataSourceConfig_duplicateSid,
				ExpectError: regexache.MustCompile(`duplicate Sid \(S1\)`),
			},
		},
	})
}

func TestAccIAMPolicyDocumentsDataSource_oversizedStatement(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPolicyDocumentsDataSourceConfig_split(100),
				ExpectError: regexache.MustCompile(fmt.Sprintf(`statement 0 \(Sid "S1"\) of the merged policy document is %d characters on its own, exceeding\s+max_policy_size \(100\)`, 128+len(acctest.Partition())-len("aws"))),
			},
		},
	})
}

func TestAccIAMPolicyDocumentsDataSource_variables(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_policy_documents.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentsDataSourceConfig_variables("2012-10-17"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "minified_json.0", fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"arn:%s:s3:::bucket/home/${aws:username}/*","Principal":{"AWS":"*"},"Condition":{"StringLike":{"s3:prefix":"home/${aws:username}/"}}}]}`, acctest.Partition())),
				),
			},
			{
				Config:      testAccPolicyDocumentsDataSourceConfig_variables("2008-10-17"),
				ExpectError: regexache.MustCompile(`found &\{ sequence in \(.+\), which is not supported in document version\s+2008-10-17`),
			},
		},
	})
}

const testAccPolicyDocumentsDataSourceConfig_basic = `
data "aws_partition" "current" {}

data "aws_iam_policy_documents" "test" {
  statement {
    sid       = "S1"
    actions   = ["s3:GetObject", "s3:PutObject"]
    resources = ["arn:${data.aws_partition.current.partition}:s3:::bucket1/*"]
  }
}
`

func testAccPolicyDocumentsDataSourceConfig_split(maxPolicySize int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_documents" "test" {
  max_policy_size = %[1]d

  dynamic "statement" {
    for_each = [1, 2, 3]

    content {
      sid       = "S${statement.value}"
      actions   = ["s3:GetObject"]
      resources = ["arn:${data.aws_partition.current.partition}:s3:::bucket${statement.value}/*"]
    }
  }
}
`, maxPolicySize)
}

const testAccPolicyDocumentsDataSourceConfig_sourceOverride = `
data "aws_iam_policy_documents" "test" {
  source_policy_documents = [jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid      = "Source"
      Effect   = "Allow"
      Action   = "s3:PutObject"
      Resource = "*"
    }]
  })]

  override_policy_documents = [jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid      = "Source"
      Effect   = "Deny"
      Action   = "s3:DeleteObject"
      Resource = "*"
    }]
  })]

  statement {
    sid       = "Config"
    actions   = ["s3:GetObject"]
    resources = ["*"]
  }
}
`

const testAccPolicyDocumentsDataSourceConfig_emptyString = `
data "aws_iam_policy_documents" "test" {
  source_policy_documents   = [""]
  override_policy_documents = [""]
}
`

const testAccPolicyDocumentsDataSourceConfig_duplicateSid = `
data "aws_iam_policy_documents" "test" {
  statement {
    sid       = "S1"
    actions   = ["s3:GetObject"]
    resources = ["*"]
  }

  statement {
    sid       = "S1"
    actions   = ["s3:PutObject"]
    resources = ["*"]
  }
}
`

func testAccPolicyDocumentsDataSourceConfig_variables(version string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_documents" "test" {
  version = %[1]q

  statement {
    actions   = ["s3:GetObject"]
    resources = ["arn:${data.aws_partition.current.partition}:s3:::bucket/home/&{aws:username}/*"]

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values   = ["home/&{aws:username}/"]
    }
  }
}
`, version)
}
