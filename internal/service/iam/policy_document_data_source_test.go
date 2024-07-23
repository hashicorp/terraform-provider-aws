// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMPolicyDocumentDataSource_basic(t *testing.T) {
	// This really ought to be able to be a unit test rather than an
	// acceptance test, but just instantiating the AWS provider requires
	// some AWS API calls, and so this needs valid AWS credentials to work.
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.test", names.AttrJSON,
						testAccPolicyDocumentExpectedJSON(),
					),
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.test", "minified_json",
						testAccPolicyDocumentExpectedJSONMinified(),
					),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_singleConditionValue(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_policy_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentDataSourceConfig_singleConditionValue,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrEquivalentJSON(dataSourceName, names.AttrJSON, testAccPolicyDocumentConfig_SingleConditionValue_ExpectedJSON),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_multipleConditionKeys(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_policy_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentDataSourceConfig_multipleConditionKeys,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrEquivalentJSON(dataSourceName, names.AttrJSON, testAccPolicyDocumentConfig_multipleConditionKeys_ExpectedJSON),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_duplicateConditionKeys(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_policy_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentDataSourceConfig_duplicateConditionKeys,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrEquivalentJSON(dataSourceName, names.AttrJSON, testAccPolicyDocumentConfig_duplicateConditionKeys_ExpectedJSON),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_conditionWithBoolValue(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentConfig_conditionWithBoolValue,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrEquivalentJSON("data.aws_iam_policy_document.test", names.AttrJSON,
						testAccPolicyDocumentConditionWithBoolValueExpectedJSON(),
					),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_source(t *testing.T) {
	// This really ought to be able to be a unit test rather than an
	// acceptance test, but just instantiating the AWS provider requires
	// some AWS API calls, and so this needs valid AWS credentials to work.
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentDataSourceConfig_deprecated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.test_source", names.AttrJSON,
						testAccPolicyDocumentSourceExpectedJSON(),
					),
				),
			},
			{
				Config: testAccPolicyDocumentDataSourceConfig_blankDeprecated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.test_source_blank", names.AttrJSON,
						testAccPolicyDocumentSourceBlankExpectedJSON,
					),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_sourceList(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentDataSourceConfig_list,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.test_source_list", names.AttrJSON,
						testAccPolicyDocumentSourceListExpectedJSON,
					),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_sourceConflicting(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentDataSourceConfig_conflictingDeprecated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.test_source_conflicting", names.AttrJSON,
						testAccPolicyDocumentSourceConflictingExpectedJSON,
					),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_sourceListConflicting(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPolicyDocumentDataSourceConfig_listConflicting,
				ExpectError: regexache.MustCompile(`duplicate Sid (.*?)`),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_override(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentDataSourceConfig_overrideDeprecated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.test_override", names.AttrJSON,
						testAccPolicyDocumentOverrideExpectedJSON,
					),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_overrideList(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentDataSourceConfig_overrideList,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.test_override_list", names.AttrJSON,
						testAccPolicyDocumentOverrideListExpectedJSON,
					),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_noStatementMerge(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentDataSourceConfig_noStatementMergeDeprecated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.yak_politik", names.AttrJSON,
						testAccPolicyDocumentNoStatementMergeExpectedJSON,
					),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_noStatementOverride(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentDataSourceConfig_noStatementOverrideDeprecated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.yak_politik", names.AttrJSON,
						testAccPolicyDocumentNoStatementOverrideExpectedJSON,
					),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_duplicateSid(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPolicyDocumentDataSourceConfig_duplicateSid,
				ExpectError: regexache.MustCompile(`duplicate Sid`),
			},
			{
				Config: testAccPolicyDocumentDataSourceConfig_duplicateBlankSid,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.test", names.AttrJSON,
						testAccPolicyDocumentDuplicateBlankSidExpectedJSON,
					),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_sourcePolicyValidJSON(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPolicyDocumentDataSourceConfig_invalidJSON,
				ExpectError: regexache.MustCompile(`"source_policy_documents.0" contains an invalid JSON: unexpected end of JSON input`),
			},
			{
				Config: testAccPolicyDocumentDataSourceConfig_emptyString,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.test", names.AttrJSON,
						testAccPolicyDocumentExpectedJSONNoStatement,
					),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_overridePolicyDocumentValidJSON(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPolicyDocumentDataSourceConfig_overridePolicyDocument_invalidJSON,
				ExpectError: regexache.MustCompile(`"override_policy_documents.0" contains an invalid JSON: unexpected end of JSON input`),
			},
			{
				Config: testAccPolicyDocumentDataSourceConfig_overridePolicyDocument_emptyString,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.test", names.AttrJSON,
						testAccPolicyDocumentExpectedJSONNoStatement,
					),
				),
			},
			{
				Config: testAccPolicyDocumentDataSourceConfig_sourcePolicyDocument_emptyString,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/10777
func TestAccIAMPolicyDocumentDataSource_StatementPrincipalIdentifiers_stringAndSlice(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_policy_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentDataSourceConfig_statementPrincipalIdentifiersStringAndSlice,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrJSON, testAccPolicyDocumentExpectedJSONStatementPrincipalIdentifiersStringAndSlice()),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/10777
func TestAccIAMPolicyDocumentDataSource_StatementPrincipalIdentifiers_multiplePrincipals(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_policy_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartition(t, names.StandardPartitionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentDataSourceConfig_statementPrincipalIdentifiersMultiplePrincipals,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrJSON, testAccPolicyDocumentExpectedJSONStatementPrincipalIdentifiersMultiplePrincipals()),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_StatementPrincipalIdentifiers_multiplePrincipalsGov(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_policy_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartition(t, names.USGovCloudPartitionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentDataSourceConfig_statementPrincipalIdentifiersMultiplePrincipals,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrJSON, testAccPolicyDocumentExpectedJSONStatementPrincipalIdentifiersMultiplePrincipalsGov()),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_version20081017(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPolicyDocumentDataSourceConfig_version20081017ConversionCondition,
				ExpectError: regexache.MustCompile(`found \&\{ sequence in \(.+\), which is not supported in document version 2008-10-17`),
			},
			{
				Config:      testAccPolicyDocumentDataSourceConfig_version20081017ConversionNotPrincipals,
				ExpectError: regexache.MustCompile(`found \&\{ sequence in \(.+\), which is not supported in document version 2008-10-17`),
			},
			{
				Config:      testAccPolicyDocumentDataSourceConfig_version20081017ConversionNotRe,
				ExpectError: regexache.MustCompile(`found \&\{ sequence in \(.+\), which is not supported in document version 2008-10-17`),
			},
			{
				Config:      testAccPolicyDocumentDataSourceConfig_version20081017ConversionPrincipal,
				ExpectError: regexache.MustCompile(`found \&\{ sequence in \(.+\), which is not supported in document version 2008-10-17`),
			},
			{
				Config:      testAccPolicyDocumentDataSourceConfig_version20081017ConversionRe,
				ExpectError: regexache.MustCompile(`found \&\{ sequence in \(.+\), which is not supported in document version 2008-10-17`),
			},
			{
				Config: testAccPolicyDocumentDataSourceConfig_version20081017,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.test", names.AttrJSON, testAccPolicyDocumentVersion20081017ExpectedJSONDataSourceConfig),
				),
			},
		},
	})
}

var testAccPolicyDocumentDataSourceConfig_basic = `
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  policy_id = "policy_id"

  statement {
    sid = "1"
    actions = [
      "s3:ListAllMyBuckets",
      "s3:GetBucketLocation",
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::*",
    ]
  }

  statement {
    actions = [
      "s3:ListBucket",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::foo",
    ]

    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values = [
        "home/",
        "",
        "home/&{aws:username}/",
      ]
    }

    not_principals {
      type        = "AWS"
      identifiers = ["arn:blahblah:example"]
    }
  }

  statement {
    actions = [
      "s3:*",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::foo/home/&{aws:username}",
      "arn:${data.aws_partition.current.partition}:s3:::foo/home/&{aws:username}/*",
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:blahblah:example"]
    }
  }

  statement {
    effect        = "Deny"
    not_actions   = ["s3:*"]
    not_resources = ["arn:${data.aws_partition.current.partition}:s3:::*"]
  }

  # Normalization of wildcard principals

  statement {
    effect  = "Allow"
    actions = ["kinesis:*"]

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
  }

  statement {
    effect  = "Allow"
    actions = ["firehose:*"]

    principals {
      type        = "*"
      identifiers = ["*"]
    }
  }
}
`

func testAccPolicyDocumentExpectedJSON() string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Id": "policy_id",
  "Statement": [
    {
      "Sid": "1",
      "Effect": "Allow",
      "Action": [
        "s3:ListAllMyBuckets",
        "s3:GetBucketLocation"
      ],
      "Resource": "arn:%[1]s:s3:::*"
    },
    {
      "Effect": "Allow",
      "Action": "s3:ListBucket",
      "Resource": "arn:%[1]s:s3:::foo",
      "NotPrincipal": {
        "AWS": "arn:blahblah:example"
      },
      "Condition": {
        "StringLike": {
          "s3:prefix": [
            "home/",
            "",
            "home/${aws:username}/"
          ]
        }
      }
    },
    {
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": [
        "arn:%[1]s:s3:::foo/home/${aws:username}/*",
        "arn:%[1]s:s3:::foo/home/${aws:username}"
      ],
      "Principal": {
        "AWS": "arn:blahblah:example"
      }
    },
    {
      "Effect": "Deny",
      "NotAction": "s3:*",
      "NotResource": "arn:%[1]s:s3:::*"
    },
    {
      "Effect": "Allow",
      "Action": "kinesis:*",
      "Principal": {
        "AWS": "*"
      }
    },
    {
      "Effect": "Allow",
      "Action": "firehose:*",
      "Principal": "*"
    }
  ]
}`, acctest.Partition())
}

func testAccPolicyDocumentExpectedJSONMinified() string {
	return fmt.Sprintf(`{"Version":"2012-10-17","Id":"policy_id","Statement":[{"Sid":"1","Effect":"Allow","Action":["s3:ListAllMyBuckets","s3:GetBucketLocation"],"Resource":"arn:%[1]s:s3:::*"},{"Effect":"Allow","Action":"s3:ListBucket","Resource":"arn:%[1]s:s3:::foo","NotPrincipal":{"AWS":"arn:blahblah:example"},"Condition":{"StringLike":{"s3:prefix":["home/","","home/${aws:username}/"]}}},{"Effect":"Allow","Action":"s3:*","Resource":["arn:%[1]s:s3:::foo/home/${aws:username}/*","arn:%[1]s:s3:::foo/home/${aws:username}"],"Principal":{"AWS":"arn:blahblah:example"}},{"Effect":"Deny","NotAction":"s3:*","NotResource":"arn:%[1]s:s3:::*"},{"Effect":"Allow","Action":"kinesis:*","Principal":{"AWS":"*"}},{"Effect":"Allow","Action":"firehose:*","Principal":"*"}]}`, acctest.Partition())
}

const testAccPolicyDocumentDataSourceConfig_singleConditionValue = `
data "aws_iam_policy_document" "test" {
  statement {
    effect = "Deny"

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    actions = ["elasticfilesystem:*"]

    resources = ["*"]

    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = ["false"]
    }
  }
}
`

const testAccPolicyDocumentConfig_SingleConditionValue_ExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Deny",
      "Action": "elasticfilesystem:*",
      "Resource": "*",
      "Principal": {
        "AWS": "*"
      },
      "Condition": {
        "Bool": {
          "aws:SecureTransport": "false"
        }
      }
    }
  ]
}`

const testAccPolicyDocumentDataSourceConfig_multipleConditionKeys = `
data "aws_iam_policy_document" "test" {
  statement {
    sid = "AWSCloudTrailWrite20150319"

    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["cloudtrail.amazonaws.com"]
    }

    actions = ["s3:PutObject"]

    resources = ["*"]

    condition {
      test     = "StringEquals"
      variable = "s3:x-amz-acl"
      values   = ["bucket-owner-full-control"]
    }
    condition {
      test     = "StringEquals"
      variable = "aws:SourceArn"
      values   = ["some-other-value"]
    }
  }
}
`

var testAccPolicyDocumentConfig_multipleConditionKeys_ExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailWrite20150319",
      "Effect": "Allow",
      "Action": "s3:PutObject",
      "Resource": "*",
      "Principal": {
        "Service": "cloudtrail.amazonaws.com"
      },
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control",
          "aws:SourceArn": "some-other-value"
        }
      }
    }
  ]
}
`

const testAccPolicyDocumentDataSourceConfig_duplicateConditionKeys = `
data "aws_iam_policy_document" "test" {
  statement {
    sid = "DuplicateConditionTest"

    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["cloudtrail.amazonaws.com"]
    }

    actions = ["s3:PutObject"]

    resources = ["*"]

    condition {
      test     = "StringEquals"
      variable = "s3:prefix"
      values   = ["one/", "two/"]
    }
    condition {
      test     = "StringEquals"
      variable = "s3:prefix"
      values   = ["three/"]
    }
  }
}
`

const testAccPolicyDocumentConfig_duplicateConditionKeys_ExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "DuplicateConditionTest",
      "Effect": "Allow",
      "Action": "s3:PutObject",
      "Resource": "*",
      "Principal": {
        "Service": "cloudtrail.amazonaws.com"
      },
      "Condition": {
        "StringEquals": {
          "s3:prefix": ["one/", "two/", "three/"]
        }
      }
    }
  ]
}
`

var testAccPolicyDocumentDataSourceConfig_deprecated = `
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  policy_id = "policy_id"

  statement {
    sid = "1"
    actions = [
      "s3:ListAllMyBuckets",
      "s3:GetBucketLocation",
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::*",
    ]
  }

  statement {
    actions = [
      "s3:ListBucket",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::foo",
    ]

    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values = [
        "home/",
        "home/&{aws:username}/",
      ]
    }

    not_principals {
      type        = "AWS"
      identifiers = ["arn:blahblah:example"]
    }
  }

  statement {
    actions = [
      "s3:*",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::foo/home/&{aws:username}",
      "arn:${data.aws_partition.current.partition}:s3:::foo/home/&{aws:username}/*",
    ]

    principals {
      type = "AWS"
      identifiers = [
        "arn:blahblah:example",
        "arn:blahblahblah:example",
      ]
    }
  }

  statement {
    effect        = "Deny"
    not_actions   = ["s3:*"]
    not_resources = ["arn:${data.aws_partition.current.partition}:s3:::*"]
  }

  # Normalization of wildcard principals

  statement {
    effect  = "Allow"
    actions = ["kinesis:*"]

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
  }

  statement {
    effect  = "Allow"
    actions = ["firehose:*"]

    principals {
      type        = "*"
      identifiers = ["*"]
    }
  }
}

data "aws_iam_policy_document" "test_source" {
  source_policy_documents = [data.aws_iam_policy_document.test.json]

  statement {
    sid       = "SourceJSONTest1"
    actions   = ["*"]
    resources = ["*"]
  }
}
`

func testAccPolicyDocumentSourceExpectedJSON() string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Id": "policy_id",
  "Statement": [
    {
      "Sid": "1",
      "Effect": "Allow",
      "Action": [
        "s3:ListAllMyBuckets",
        "s3:GetBucketLocation"
      ],
      "Resource": "arn:%[1]s:s3:::*"
    },
    {
      "Effect": "Allow",
      "Action": "s3:ListBucket",
      "Resource": "arn:%[1]s:s3:::foo",
      "NotPrincipal": {
        "AWS": "arn:blahblah:example"
      },
      "Condition": {
        "StringLike": {
          "s3:prefix": [
            "home/",
            "home/${aws:username}/"
          ]
        }
      }
    },
    {
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": [
        "arn:%[1]s:s3:::foo/home/${aws:username}/*",
        "arn:%[1]s:s3:::foo/home/${aws:username}"
      ],
      "Principal": {
        "AWS": [
          "arn:blahblahblah:example",
          "arn:blahblah:example"
        ]
      }
    },
    {
      "Effect": "Deny",
      "NotAction": "s3:*",
      "NotResource": "arn:%[1]s:s3:::*"
    },
    {
      "Effect": "Allow",
      "Action": "kinesis:*",
      "Principal": {
        "AWS": "*"
      }
    },
    {
      "Effect": "Allow",
      "Action": "firehose:*",
      "Principal": "*"
    },
    {
      "Sid": "SourceJSONTest1",
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*"
    }
  ]
}`, acctest.Partition())
}

var testAccPolicyDocumentDataSourceConfig_list = `
data "aws_iam_policy_document" "policy_a" {
  statement {
    sid     = ""
    effect  = "Allow"
    actions = ["foo:ActionOne"]
  }

  statement {
    sid     = "validSidOne"
    effect  = "Allow"
    actions = ["bar:ActionOne"]
  }
}

data "aws_iam_policy_document" "policy_b" {
  statement {
    sid     = "validSidTwo"
    effect  = "Deny"
    actions = ["foo:ActionTwo"]
  }
}

data "aws_iam_policy_document" "policy_c" {
  statement {
    sid     = ""
    effect  = "Allow"
    actions = ["bar:ActionTwo"]
  }
}

data "aws_iam_policy_document" "test_source_list" {
  version = "2012-10-17"

  source_policy_documents = [
    data.aws_iam_policy_document.policy_a.json,
    data.aws_iam_policy_document.policy_b.json,
    data.aws_iam_policy_document.policy_c.json
  ]
}
`
var testAccPolicyDocumentSourceListExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "foo:ActionOne"
    },
    {
      "Sid": "validSidOne",
      "Effect": "Allow",
      "Action": "bar:ActionOne"
    },
    {
      "Sid": "validSidTwo",
      "Effect": "Deny",
      "Action": "foo:ActionTwo"
    },
    {
      "Effect": "Allow",
      "Action": "bar:ActionTwo"
    }
  ]
}`

var testAccPolicyDocumentDataSourceConfig_blankDeprecated = `
data "aws_iam_policy_document" "test_source_blank" {
  source_policy_documents = [""]

  statement {
    sid       = "SourceJSONTest2"
    actions   = ["*"]
    resources = ["*"]
  }
}
`

var testAccPolicyDocumentSourceBlankExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "SourceJSONTest2",
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*"
    }
  ]
}`

var testAccPolicyDocumentDataSourceConfig_conflictingDeprecated = `
data "aws_iam_policy_document" "test_source" {
  statement {
    sid       = "SourceJSONTestConflicting"
    actions   = ["iam:*"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "test_source_conflicting" {
  source_policy_documents = [data.aws_iam_policy_document.test_source.json]

  statement {
    sid       = "SourceJSONTestConflicting"
    actions   = ["*"]
    resources = ["*"]
  }
}
`

var testAccPolicyDocumentSourceConflictingExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "SourceJSONTestConflicting",
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*"
    }
  ]
}`

var testAccPolicyDocumentDataSourceConfig_listConflicting = `
data "aws_iam_policy_document" "policy_a" {
  statement {
    sid     = ""
    effect  = "Allow"
    actions = ["foo:ActionOne"]
  }

  statement {
    sid     = "conflictSid"
    effect  = "Allow"
    actions = ["bar:ActionOne"]
  }
}

data "aws_iam_policy_document" "policy_b" {
  statement {
    sid     = "validSid"
    effect  = "Deny"
    actions = ["foo:ActionTwo"]
  }
}

data "aws_iam_policy_document" "policy_c" {
  statement {
    sid     = "conflictSid"
    effect  = "Allow"
    actions = ["bar:ActionTwo"]
  }
}

data "aws_iam_policy_document" "test_source_list_conflicting" {
  version = "2012-10-17"

  source_policy_documents = [
    data.aws_iam_policy_document.policy_a.json,
    data.aws_iam_policy_document.policy_b.json,
    data.aws_iam_policy_document.policy_c.json
  ]
}
`

var testAccPolicyDocumentDataSourceConfig_overrideDeprecated = `
data "aws_partition" "current" {}

data "aws_iam_policy_document" "override" {
  statement {
    sid = "SidToOverwrite"

    actions   = ["s3:*"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "test_override" {
  override_policy_documents = [data.aws_iam_policy_document.override.json]

  statement {
    actions   = ["ec2:*"]
    resources = ["*"]
  }

  statement {
    sid = "SidToOverwrite"

    actions = ["s3:*"]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::somebucket",
      "arn:${data.aws_partition.current.partition}:s3:::somebucket/*",
    ]
  }
}
`

var testAccPolicyDocumentOverrideExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "ec2:*",
      "Resource": "*"
    },
    {
      "Sid": "SidToOverwrite",
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": "*"
    }
  ]
}`

var testAccPolicyDocumentDataSourceConfig_overrideList = `
data "aws_iam_policy_document" "policy_a" {
  statement {
    sid     = ""
    effect  = "Allow"
    actions = ["foo:ActionOne"]
  }

  statement {
    sid     = "overrideSid"
    effect  = "Allow"
    actions = ["bar:ActionOne"]
  }
}

data "aws_iam_policy_document" "policy_b" {
  statement {
    sid     = "validSid"
    effect  = "Deny"
    actions = ["foo:ActionTwo"]
  }
}

data "aws_iam_policy_document" "policy_c" {
  statement {
    sid     = "overrideSid"
    effect  = "Deny"
    actions = ["bar:ActionOne"]
  }
}

data "aws_iam_policy_document" "test_override_list" {
  version = "2012-10-17"

  override_policy_documents = [
    data.aws_iam_policy_document.policy_a.json,
    data.aws_iam_policy_document.policy_b.json,
    data.aws_iam_policy_document.policy_c.json
  ]
}
`

var testAccPolicyDocumentOverrideListExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "foo:ActionOne"
    },
    {
      "Sid": "overrideSid",
      "Effect": "Deny",
      "Action": "bar:ActionOne"
    },
    {
      "Sid": "validSid",
      "Effect": "Deny",
      "Action": "foo:ActionTwo"
    }
  ]
}`

var testAccPolicyDocumentDataSourceConfig_noStatementMergeDeprecated = `
data "aws_iam_policy_document" "source" {
  statement {
    sid       = ""
    actions   = ["ec2:DescribeAccountAttributes"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "override" {
  statement {
    sid       = "OverridePlaceholder"
    actions   = ["s3:GetObject"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "yak_politik" {
  source_policy_documents   = [data.aws_iam_policy_document.source.json]
  override_policy_documents = [data.aws_iam_policy_document.override.json]
}
`

var testAccPolicyDocumentNoStatementMergeExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "ec2:DescribeAccountAttributes",
      "Resource": "*"
    },
    {
      "Sid": "OverridePlaceholder",
      "Effect": "Allow",
      "Action": "s3:GetObject",
      "Resource": "*"
    }
  ]
}`

var testAccPolicyDocumentDataSourceConfig_noStatementOverrideDeprecated = `
data "aws_iam_policy_document" "source" {
  statement {
    sid       = "OverridePlaceholder"
    actions   = ["ec2:DescribeAccountAttributes"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "override" {
  statement {
    sid       = "OverridePlaceholder"
    actions   = ["s3:GetObject"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "yak_politik" {
  source_policy_documents   = [data.aws_iam_policy_document.source.json]
  override_policy_documents = [data.aws_iam_policy_document.override.json]
}
`

var testAccPolicyDocumentNoStatementOverrideExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "OverridePlaceholder",
      "Effect": "Allow",
      "Action": "s3:GetObject",
      "Resource": "*"
    }
  ]
}`

var testAccPolicyDocumentDataSourceConfig_duplicateSid = `
data "aws_iam_policy_document" "test" {
  statement {
    sid       = "1"
    effect    = "Allow"
    actions   = ["ec2:DescribeAccountAttributes"]
    resources = ["*"]
  }

  statement {
    sid       = "1"
    effect    = "Allow"
    actions   = ["s3:GetObject"]
    resources = ["*"]
  }
}
`

var testAccPolicyDocumentDataSourceConfig_duplicateBlankSid = `
data "aws_iam_policy_document" "test" {
  statement {
    sid       = ""
    effect    = "Allow"
    actions   = ["ec2:DescribeAccountAttributes"]
    resources = ["*"]
  }

  statement {
    sid       = ""
    effect    = "Allow"
    actions   = ["s3:GetObject"]
    resources = ["*"]
  }
}
`

var testAccPolicyDocumentDuplicateBlankSidExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "ec2:DescribeAccountAttributes",
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": "s3:GetObject",
      "Resource": "*"
    }
  ]
}`

const testAccPolicyDocumentDataSourceConfig_version20081017 = `
data "aws_iam_policy_document" "test" {
  version = "2008-10-17"

  statement {
    actions   = ["ec2:*"]
    resources = ["*"]
  }
}
`

const testAccPolicyDocumentVersion20081017ExpectedJSONDataSourceConfig = `{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "ec2:*",
      "Resource": "*"
    }
  ]
}`

const testAccPolicyDocumentDataSourceConfig_version20081017ConversionCondition = `
data "aws_iam_policy_document" "test" {
  version = "2008-10-17"

  statement {
    actions   = ["*"]
    resources = ["*"]

    condition {
      test = "StringLike"
      values = [
        "home/",
        "home/&{aws:username}/",
      ]
      variable = "s3:prefix"
    }
  }
}
`

const testAccPolicyDocumentDataSourceConfig_version20081017ConversionNotPrincipals = `
data "aws_iam_policy_document" "test" {
  version = "2008-10-17"

  statement {
    actions   = ["*"]
    resources = ["*"]

    not_principals {
      identifiers = ["&{aws:username}"]
      type        = "AWS"
    }
  }
}
`

const testAccPolicyDocumentDataSourceConfig_version20081017ConversionNotRe = `
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  version = "2008-10-17"

  statement {
    actions       = ["*"]
    not_resources = ["arn:${data.aws_partition.current.partition}:s3:::foo/home/&{aws:username}", ]
  }
}
`

const testAccPolicyDocumentDataSourceConfig_version20081017ConversionPrincipal = `
data "aws_iam_policy_document" "test" {
  version = "2008-10-17"

  statement {
    actions   = ["*"]
    resources = ["*"]

    principals {
      identifiers = ["&{aws:username}"]
      type        = "AWS"
    }
  }
}
`

const testAccPolicyDocumentDataSourceConfig_version20081017ConversionRe = `
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  version = "2008-10-17"

  statement {
    actions   = ["*"]
    resources = ["arn:${data.aws_partition.current.partition}:s3:::foo/home/&{aws:username}", ]
  }
}
`

var testAccPolicyDocumentDataSourceConfig_statementPrincipalIdentifiersStringAndSlice = `
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["*"]
    resources = ["*"]
    sid       = "StatementPrincipalIdentifiersStringAndSlice"

    principals {
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::111111111111:root"]
      type        = "AWS"
    }

    principals {
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::222222222222:root", "arn:${data.aws_partition.current.partition}:iam::333333333333:root"]
      type        = "AWS"
    }
  }
}
`

const testAccPolicyDocumentConfig_conditionWithBoolValue = `
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  source_policy_documents = [<<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "RestrictAccessToSpecialTag",
            "Effect": "Deny",
            "Action": [
                "ec2:CreateTags",
                "ec2:DeleteTags"
            ],
            "Resource": "arn:${data.aws_partition.current.partition}:ec2:*:*:vpc/*",
            "Condition": {
                "Null": {
                    "aws:ResourceTag/SpecialTag": false
                },
                "StringLike": {
                    "aws:ResourceAccount": [
                        "123456"
                    ],
                    "aws:PrincipalArn": "arn:${data.aws_partition.current.partition}:iam::*:role/AWSAFTExecution"
                }
            }
        }
    ]
}
EOF
  ]
}
`

func testAccPolicyDocumentConditionWithBoolValueExpectedJSON() string {
	return fmt.Sprintf(`{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "RestrictAccessToSpecialTag",
            "Effect": "Deny",
            "Action": [
                "ec2:CreateTags",
                "ec2:DeleteTags"
            ],
            "Resource": "arn:%[1]s:ec2:*:*:vpc/*",
            "Condition": {
                "Null": {
                    "aws:ResourceTag/SpecialTag": "false"
                },
                "StringLike": {
                    "aws:ResourceAccount": "123456",
                    "aws:PrincipalArn": "arn:%[1]s:iam::*:role/AWSAFTExecution"
                }
            }
        }
    ]
  }`, acctest.Partition())
}

func testAccPolicyDocumentExpectedJSONStatementPrincipalIdentifiersStringAndSlice() string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "StatementPrincipalIdentifiersStringAndSlice",
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*",
      "Principal": {
        "AWS": [
          "arn:%[1]s:iam::111111111111:root",
          "arn:%[1]s:iam::333333333333:root",
          "arn:%[1]s:iam::222222222222:root"
        ]
      }
    }
  ]
}`, acctest.Partition())
}

var testAccPolicyDocumentDataSourceConfig_statementPrincipalIdentifiersMultiplePrincipals = `
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["*"]
    resources = ["*"]
    sid       = "StatementPrincipalIdentifiersStringAndSlice"

    principals {
      identifiers = [
        "arn:${data.aws_partition.current.partition}:iam::111111111111:root",
        "arn:${data.aws_partition.current.partition}:iam::222222222222:root",
      ]
      type = "AWS"
    }

    principals {
      identifiers = [
        "arn:${data.aws_partition.current.partition}:iam::333333333333:root",
      ]
      type = "AWS"
    }

    principals {
      identifiers = [
        "arn:${data.aws_partition.current.partition}:iam::444444444444:root",
      ]
      type = "AWS"
    }
  }
}
`

var testAccPolicyDocumentDataSourceConfig_emptyString = `
data "aws_iam_policy_document" "test" {
  source_policy_documents = [""]
}
`

var testAccPolicyDocumentDataSourceConfig_invalidJSON = `
data "aws_iam_policy_document" "test" {
  source_policy_documents = ["{"]
}
`

var testAccPolicyDocumentExpectedJSONNoStatement = `{
  "Version": "2012-10-17"
}`

func testAccPolicyDocumentExpectedJSONStatementPrincipalIdentifiersMultiplePrincipals() string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "StatementPrincipalIdentifiersStringAndSlice",
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*",
      "Principal": {
        "AWS": [
          "arn:%[1]s:iam::333333333333:root",
          "arn:%[1]s:iam::444444444444:root",
          "arn:%[1]s:iam::222222222222:root",
          "arn:%[1]s:iam::111111111111:root"
        ]
      }
    }
  ]
}`, acctest.Partition())
}

func testAccPolicyDocumentExpectedJSONStatementPrincipalIdentifiersMultiplePrincipalsGov() string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "StatementPrincipalIdentifiersStringAndSlice",
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*",
      "Principal": {
        "AWS": [
          "arn:%[1]s:iam::333333333333:root",
          "arn:%[1]s:iam::222222222222:root",
          "arn:%[1]s:iam::111111111111:root",
          "arn:%[1]s:iam::444444444444:root"
        ]
      }
    }
  ]
}`, acctest.Partition())
}

var testAccPolicyDocumentDataSourceConfig_overridePolicyDocument_emptyString = `
data "aws_iam_policy_document" "test" {
  override_policy_documents = [""]
}
`

var testAccPolicyDocumentDataSourceConfig_overridePolicyDocument_invalidJSON = `
data "aws_iam_policy_document" "test" {
  override_policy_documents = ["{"]
}
`

var testAccPolicyDocumentDataSourceConfig_sourcePolicyDocument_emptyString = `
variable "additional_policy_statements" {
  type        = string
  description = "additional policy statements that can be added to the role's policy document"
  default     = ""
}

data "aws_iam_policy_document" "assume_role_policy" {
  source_policy_documents = [
    data.aws_iam_policy_document.partial.json,
    var.additional_policy_statements
  ]
}

data "aws_iam_policy_document" "partial" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}
`
