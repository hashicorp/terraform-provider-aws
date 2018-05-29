---
layout: "aws"
page_title: "AWS: aws_organizations_organization"
sidebar_current: "docs-aws-datasource-organizations_organization"
description: |-
  Get information on the organization that the user's account belongs to.
---

# Data Source: aws_organizations_organization

Use this data source to get details on the joined organization.

~> **NOTE:** The AWS Account must be a member of an AWS Organization.

## Example Usage

```hcl
data "aws_organizations_organization" "org" {}

resource "aws_s3_bucket" "example" {
  bucket = "my_tf_test_bucket"
}

data "aws_iam_policy_document" "bucket" {
  statement {
    actions = [
      "s3:*",
    ]

    resources = ["arn:aws:s3:::${aws_s3_bucket.example.id}/*"]

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    condition {
      test     = "StringEquals"
      variable = "aws:PrincipalOrgID"

      values = [
        "${data.aws_organizations_organization.org.id}",
      ]
    }
  }
}

resource "aws_s3_bucket_policy" "example" {
  bucket = "${aws_s3_bucket.example.id}"
  policy = "${data.aws_iam_policy_document.bucket.json}"
}
```

## Argument Reference

There are no arguments available for this data source.

## Attributes Reference

`id` is set to the ID of the Organization. In addition, the following attributes
are exported:

* `arn` - ARN of the organization
* `master_account_arn` - ARN of the master account
* `master_account_email` - Email address of the master account
* `master_account_id` - Identifier of the master account
* `feature_set` - Specifies the functionality that currently is available to the
  organization.If set to "ALL", then all features are enabled and policies can be
  applied to accounts in the organization. If set to "CONSOLIDATED_BILLING", then
  only consolidated billing functionality is available.
