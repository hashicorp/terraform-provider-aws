---
subcategory: "CloudWatch Omni"
layout: "aws"
page_title: "AWS: aws_cloudwatchomni_space"
description: |-
  Manages a CloudWatch Omni Space.
---

# Resource: aws_cloudwatchomni_space

Manages a CloudWatch Omni Space. A space is the container that holds your observability data within a CloudWatch Omni domain.

~> **NOTE:** CloudWatch Omni allows only **one space per AWS account per region**. Creating a second space in the same region returns `ConflictException`. A space in a different region is permitted, and a space does not have to be in the same region as its domain.

~> **NOTE:** CloudWatch Omni is not available in every region. In unsupported regions the API returns `AccessDeniedException`.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_role" "example" {
  name = "cloudwatch-omni-data-access"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "cloudwatch.amazonaws.com"
      }
    }]
  })
}

resource "aws_cloudwatchomni_domain" "example" {
  name               = "example"
  identity_providers = ["IAM"]
}

resource "aws_cloudwatchomni_space" "example" {
  name                 = "example"
  domain_id            = aws_cloudwatchomni_domain.example.domain_id
  data_access_role_arn = aws_iam_role.example.arn
}
```

### AgentCore Online Evaluation

Supply `agent_core_evaluation_role_arn` to let the space run AgentCore online evaluation.

```terraform
resource "aws_cloudwatchomni_space" "example" {
  name                           = "example"
  domain_id                      = aws_cloudwatchomni_domain.example.domain_id
  data_access_role_arn           = aws_iam_role.data_access.arn
  agent_core_evaluation_role_arn = aws_iam_role.agent_core.arn
}
```

### Customer Managed Encryption

Omitting `encryption_configuration` leaves the space encrypted with an AWS owned key. To use a customer managed key, the key policy must grant the `cloudwatch.amazonaws.com` service principal permission to use it, otherwise `CreateSpace` fails with `AccessDeniedException`.

```terraform
data "aws_caller_identity" "current" {}

resource "aws_kms_key" "example" {
  description = "CloudWatch Omni Space encryption"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "EnableIAMUserPermissions"
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "AllowCloudWatchOmni"
        Effect = "Allow"
        Principal = {
          Service = "cloudwatch.amazonaws.com"
        }
        Action = [
          "kms:Decrypt",
          "kms:DescribeKey",
          "kms:Encrypt",
          "kms:GenerateDataKey*",
          "kms:ReEncrypt*",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_cloudwatchomni_space" "example" {
  name                 = "example"
  domain_id            = aws_cloudwatchomni_domain.example.domain_id
  data_access_role_arn = aws_iam_role.example.arn

  encryption_configuration {
    encryption_strategy = "CUSTOMER_MANAGED"
    kms_key_arn         = aws_kms_key.example.arn
  }
}
```

## Argument Reference

The following arguments are required:

* `data_access_role_arn` - (Required) ARN of the IAM role CloudWatch Omni assumes to access data. Must be in the same account as the space. Changing this forces a new resource.
* `domain_id` - (Required) ID of the [CloudWatch Omni domain](cloudwatchomni_domain.html) the space belongs to. Changing this forces a new resource.
* `name` - (Required) Name of the space. Must be 3-64 characters of lowercase letters, numbers and hyphens, must begin and end with a letter or number, and cannot contain consecutive hyphens.

The following arguments are optional:

* `agent_core_evaluation_role_arn` - (Optional) ARN of the IAM role used by AgentCore online evaluation. Changing this forces a new resource.
* `encryption_configuration` - (Optional) Encryption of the space's data at rest. Omit for AWS owned encryption. [See below](#encryption_configuration-block).
* `region` - (Optional) Region where this resource is managed. Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `encryption_configuration` Block

* `encryption_strategy` - (Required) Encryption strategy. Valid values are `AWS_OWNED` and `CUSTOMER_MANAGED`.
* `kms_key_arn` - (Optional) ARN of the KMS key to use. Required when `encryption_strategy` is `CUSTOMER_MANAGED`.

~> **NOTE:** Specifying `encryption_strategy = "AWS_OWNED"` is equivalent to omitting the block entirely. The service reports AWS owned encryption as the absence of a customer managed key, so the block is normalized away and will not appear in state.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `domain_arn` - ARN of the domain the space belongs to.
* `space_arn` - ARN of the space.
* `space_id` - Unique ID of the space.
* `status` - Status of the space.

~> **NOTE:** This resource does not support tags. `CreateSpace` accepts tags, but CloudWatch Omni does not return them on read and provides no tag read or write operations, so tags could not be refreshed, updated, or imported.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a CloudWatch Omni Space using the `space_id`. For example:

```terraform
import {
  to = aws_cloudwatchomni_space.example
  id = "12345678-1234-1234-1234-123456789012"
}
```

Using `terraform import`, import a CloudWatch Omni Space using the `space_id`. For example:

```console
% terraform import aws_cloudwatchomni_space.example 12345678-1234-1234-1234-123456789012
```
