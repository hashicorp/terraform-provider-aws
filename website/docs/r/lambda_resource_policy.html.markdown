---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_resource_policy"
description: |-
  Manages a full IAM resource-based policy for an AWS Lambda resource.
---

# Resource: aws_lambda_resource_policy

Manages the complete IAM resource-based policy document for an AWS Lambda function, function version, or alias.

~> **Note:** `PutResourcePolicy` (used by this resource) replaces the *entire* resource-based policy on the Lambda resource, including any statements added with [`aws_lambda_permission`](/docs/providers/aws/r/lambda_permission.html). Do not use `aws_lambda_resource_policy` and `aws_lambda_permission` on the same Lambda function, version, or alias — every apply of one will overwrite statements managed by the other.

## Example Usage

### Basic Usage

```terraform
resource "aws_lambda_resource_policy" "example" {
  resource_arn = aws_lambda_function.example.arn
  policy       = data.aws_iam_policy_document.example.json
}

data "aws_iam_policy_document" "example" {
  statement {
    sid    = "AllowInvokeFromS3"
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["s3.amazonaws.com"]
    }

    actions   = ["lambda:InvokeFunction"]
    resources = [aws_lambda_function.example.arn]

    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
  }
}

data "aws_caller_identity" "current" {}
```

### Multiple Principals

```terraform
resource "aws_lambda_resource_policy" "example" {
  resource_arn = aws_lambda_function.example.arn
  policy       = data.aws_iam_policy_document.example.json
}

data "aws_iam_policy_document" "example" {
  statement {
    sid    = "AllowCrossAccountInvoke"
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["123456789012", "210987654321"]
    }

    actions   = ["lambda:InvokeFunction"]
    resources = [aws_lambda_function.example.arn]
  }

  statement {
    sid    = "AllowOrganizationInvoke"
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    actions   = ["lambda:InvokeFunction"]
    resources = [aws_lambda_function.example.arn]

    condition {
      test     = "StringEquals"
      variable = "aws:PrincipalOrgID"
      values   = ["o-1234567890"]
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `policy` - (Required) JSON-formatted resource-based policy document to attach to the Lambda resource. This replaces the entire policy on the resource. Maximum 20,480 characters.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `resource_arn` - (Required, Forces new resource) ARN of the Lambda function, function version, or function alias to attach the policy to. Can be a qualified or unqualified ARN.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `revision_id` - Unique identifier for the current revision of the policy.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_lambda_resource_policy.example
  identity = {
    "resource_arn" = "arn:aws:lambda:us-east-1:123456789012:function:example"
  }
}

resource "aws_lambda_resource_policy" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `resource_arn` (String) ARN of the Lambda function, function version, or function alias.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lambda policies using the `resource_arn`. For example:

```terraform
import {
  to = aws_lambda_resource_policy.example
  id = "arn:aws:lambda:us-east-1:123456789012:function:example"
}
```

Using `terraform import`, import Lambda policies using the `resource_arn`. For example:

```console
% terraform import aws_lambda_resource_policy.example arn:aws:lambda:us-east-1:123456789012:function:example
```
