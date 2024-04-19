---
subcategory: "Secrets Manager"
layout: "aws"
page_title: "AWS: aws_secretsmanager_secret_policy"
description: |-
  Provides a resource to manage AWS Secrets Manager secret policy
---

# Resource: aws_secretsmanager_secret_policy

Provides a resource to manage AWS Secrets Manager secret policy.

## Example Usage

### Basic

```terraform
resource "aws_secretsmanager_secret" "example" {
  name = "example"
}

data "aws_iam_policy_document" "example" {
  statement {
    sid    = "EnableAnotherAWSAccountToReadTheSecret"
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::123456789012:root"]
    }

    actions   = ["secretsmanager:GetSecretValue"]
    resources = ["*"]
  }
}

resource "aws_secretsmanager_secret_policy" "example" {
  secret_arn = aws_secretsmanager_secret.example.arn
  policy     = data.aws_iam_policy_document.example.json
}
```

## Argument Reference

The following arguments are required:

* `policy` - (Required) Valid JSON document representing a [resource policy](https://docs.aws.amazon.com/secretsmanager/latest/userguide/auth-and-access_resource-based-policies.html). For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy). Unlike `aws_secretsmanager_secret`, where `policy` can be set to `"{}"` to delete the policy, `"{}"` is not a valid policy since `policy` is required.
* `secret_arn` - (Required) Secret ARN.

The following arguments are optional:

* `block_public_policy` - (Optional) Makes an optional API call to Zelkova to validate the Resource Policy to prevent broad access to your secret.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Amazon Resource Name (ARN) of the secret.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_secretsmanager_secret_policy` using the secret Amazon Resource Name (ARN). For example:

```terraform
import {
  to = aws_secretsmanager_secret_policy.example
  id = "arn:aws:secretsmanager:us-east-1:123456789012:secret:example-123456"
}
```

Using `terraform import`, import `aws_secretsmanager_secret_policy` using the secret Amazon Resource Name (ARN). For example:

```console
% terraform import aws_secretsmanager_secret_policy.example arn:aws:secretsmanager:us-east-1:123456789012:secret:example-123456
```
