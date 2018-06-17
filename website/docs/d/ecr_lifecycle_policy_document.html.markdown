---
layout: "aws"
page_title: "AWS: aws_ecr_lifecycle_policy_document"
sidebar_current: "docs-aws-datasource-ecr-lifecycle-policy-document"
description: |-
  Generates an ECR lifecycle policy document in JSON format
---

# Data Source: aws_ecr_lifecycle_policy_document

Generates an ECR lifecycle policy document in JSON format.

This is a data source which can be used to construct a JSON representation of
an ECR lifecycle policy document, for use with `aws_ecr_lifecycle_policy`.

```hcl
data "aws_ecr_lifecycle_policy_document" "example" {
  rule {
    priority = 1
    description = "Expire images older than 14 days"

    selection {
      tag_status = "untagged"
      count_type = "sinceImagePushed"
      count_unit = "days"
      count_number = 14
    }

    action {
      type = "expire"
    }
  }

  rule {
    priority = 2
    description = "Keep last 30 images"

    selection {
      tag_status = "tagged"
      tag_prefixes = ["v"]
      count_type = "imageCountMoreThan"
      count_number = 30
    }

    action {
      type = "expire"
    }
  }
}

resource "aws_ecr_lifecycle_policy" "example" {
  name   = "example_policy"
  policy = "${data.aws_ecr_lifecycle_policy_document.example.json}"
}
```

Using this data source to generate policy documents is *optional*. It is also
valid to use literal JSON strings within your configuration, or to use the
`file` interpolation function to read a raw JSON policy document from a file.

## Argument Reference

The following arguments are supported:



## Attributes Reference

The following attribute is exported:

* `json` - The above arguments serialized as a standard JSON policy document.

## Example with Source and Override

Showing how you can use `source_json` and `override_json`

```hcl
data "aws_iam_policy_document" "source" {
  statement {
    actions   = ["ec2:*"]
    resources = ["*"]
  }

  statement {
    sid = "SidToOverwrite"

    actions   = ["s3:*"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "source_json_example" {
  source_json = "${data.aws_iam_policy_document.source.json}"

  statement {
    sid = "SidToOverwrite"

    actions = ["s3:*"]

    resources = [
      "arn:aws:s3:::somebucket",
      "arn:aws:s3:::somebucket/*",
    ]
  }
}

data "aws_iam_policy_document" "override" {
  statement {
    sid = "SidToOverwrite"

    actions   = ["s3:*"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "override_json_example" {
  override_json = "${data.aws_iam_policy_document.override.json}"

  statement {
    actions   = ["ec2:*"]
    resources = ["*"]
  }

  statement {
    sid = "SidToOverwrite"

    actions = ["s3:*"]

    resources = [
      "arn:aws:s3:::somebucket",
      "arn:aws:s3:::somebucket/*",
    ]
  }
}
```

`data.aws_iam_policy_document.source_json_example.json` will evaluate to:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "ec2:*",
      "Resource": "*"
    },
    {
      "Sid": "SidToOverwrite",
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": [
        "arn:aws:s3:::somebucket/*",
        "arn:aws:s3:::somebucket"
      ]
    }
  ]
}
```

`data.aws_iam_policy_document.override_json_example.json` will evaluate to:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
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
}
```

You can also combine `source_json` and `override_json` in the same document.
