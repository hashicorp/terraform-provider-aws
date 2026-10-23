---
subcategory: "CloudWatch Observability Admin"
layout: "aws"
page_title: "AWS: aws_observabilityadmin_dataset_integration"
description: |-
  Manages a CloudWatch Observability Admin Dataset Integration.
---

# Resource: aws_observabilityadmin_dataset_integration

Manages a CloudWatch Observability Admin Dataset Integration. A dataset integration enables CloudWatch to forward the account's telemetry to a dataset using an IAM role that you provide.

~> **NOTE:** Only one dataset integration can exist per account and Region. Creating a second one fails with a `ConflictException`. To manage an existing dataset integration, [import](#import) it.

~> **NOTE:** A successful apply does not prove that forwarding works. If the execution role's trust policy names the wrong account in `aws:SourceAccount` or `aws:SourceArn`, or the role lacks the permissions below, the integration is still created but nothing is forwarded.

Before you create the integration:

* The caller needs `iam:PassRole` on the execution role to create or update the integration.
* The execution role must be in the same account as the integration and must trust `logs.amazonaws.com`. Both trust policy conditions must name that account, as in the example below.
* The role's `logs:IntegrateWithDataset` permissions decide which log groups are forwarded. `log-group:*` forwards all log groups in the account. Log groups in the Delivery log class are never forwarded.
* If the destination dataset is encrypted with a customer managed AWS KMS key, the key policy must allow the execution role `kms:Encrypt`, `kms:Decrypt`, and `kms:GenerateDataKey`.

## Example Usage

### Basic Usage

```terraform
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_iam_role" "example" {
  name = "example-dataset-integration"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "logs.amazonaws.com"
      }
      Condition = {
        StringEquals = {
          "aws:SourceAccount" = data.aws_caller_identity.current.account_id
        }
        ArnLike = {
          "aws:SourceArn" = "arn:${data.aws_partition.current.partition}:observabilityadmin:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:dataset-integration/default"
        }
      }
    }]
  })
}

resource "aws_iam_role_policy" "example" {
  role = aws_iam_role.example.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "logs:IntegrateWithDataset"
        Resource = "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:log-group:*"
      },
      {
        Effect   = "Allow"
        Action   = "cloudwatch:PutRecords"
        Resource = "*"
      },
    ]
  })
}

resource "aws_observabilityadmin_dataset_integration" "example" {
  role_arn = aws_iam_role.example.arn

  depends_on = [aws_iam_role_policy.example]
}
```

## Argument Reference

The following arguments are required:

* `role_arn` - (Required) ARN of the IAM role that grants CloudWatch permission to access the resources needed for the dataset integration.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the dataset integration.
* `created_at` - Timestamp when the dataset integration was created, in [RFC3339 format](https://datatracker.ietf.org/doc/html/rfc3339#section-5.8).
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `updated_at` - Timestamp when the dataset integration was last updated, in [RFC3339 format](https://datatracker.ietf.org/doc/html/rfc3339#section-5.8).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_observabilityadmin_dataset_integration.example
  identity = {
    "arn" = "arn:aws:observabilityadmin:us-east-1:123456789012:dataset-integration/default"
  }
}

resource "aws_observabilityadmin_dataset_integration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) ARN of the dataset integration.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Observability Admin Dataset Integrations using the `arn`. For example:

```terraform
import {
  to = aws_observabilityadmin_dataset_integration.example
  id = "arn:aws:observabilityadmin:us-east-1:123456789012:dataset-integration/default"
}
```

Using `terraform import`, import CloudWatch Observability Admin Dataset Integrations using the `arn`. For example:

```console
% terraform import aws_observabilityadmin_dataset_integration.example arn:aws:observabilityadmin:us-east-1:123456789012:dataset-integration/default
```
