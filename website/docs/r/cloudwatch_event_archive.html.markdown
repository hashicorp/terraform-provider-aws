---
subcategory: "EventBridge"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_archive"
description: |-
  Provides an EventBridge event archive resource.
---

# Resource: aws_cloudwatch_event_archive

Provides an EventBridge event archive resource.

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.

## Example Usage

```terraform
resource "aws_cloudwatch_event_bus" "order" {
  name = "orders"
}

resource "aws_cloudwatch_event_archive" "order" {
  name             = "order-archive"
  event_source_arn = aws_cloudwatch_event_bus.order.arn
}
```

## Example Usage Optional Arguments

```terraform
resource "aws_cloudwatch_event_bus" "order" {
  name = "orders"
}

resource "aws_cloudwatch_event_archive" "order" {
  name             = "order-archive"
  description      = "Archived events from order service"
  event_source_arn = aws_cloudwatch_event_bus.order.arn
  retention_days   = 7
  event_pattern = jsonencode({
    source = ["company.team.order"]
  })
}
```

## Example Usage CMK Encryption

```terraform
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_cloudwatch_event_bus" "example" {
  name = "example"
}

resource "aws_kms_key" "example" {
  deletion_window_in_days = 7
  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "key-policy-example"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow describing of the key"
        Effect = "Allow"
        Principal = {
          Service = "events.amazonaws.com"
        },
        Action = [
          "kms:DescribeKey"
        ],
        Resource = "*"
      },
      {
        Sid    = "Allow use of the key"
        Effect = "Allow"
        Principal = {
          Service = "events.amazonaws.com"
        },
        Action = [
          "kms:GenerateDataKey",
          "kms:Decrypt",
          "kms:ReEncrypt*"
        ],
        Resource = "*"
        Condition = {
          StringEquals = {
            "kms:EncryptionContext:aws:events:event-bus:arn" = aws_cloudwatch_event_bus.example.arn
          }
        }
      }
    ]
  })
  tags = {
    EventBridgeApiDestinations = "true"
  }
}

resource "aws_cloudwatch_event_archive" "example" {
  name               = "example"
  event_source_arn   = aws_cloudwatch_event_bus.example.arn
  kms_key_identifier = aws_kms_key.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the archive. The archive name cannot exceed 48 characters.
* `event_source_arn` - (Required) ARN of the event bus associated with the archive. Only events from this event bus are sent to the archive.
* `description` - (Optional) Description for the archive.
* `event_pattern` - (Optional) Event pattern to use to filter events sent to the archive. By default, it attempts to archive every event received in the `event_source_arn`.
* `kms_key_identifier` - (Optional) Identifier of the AWS KMS customer managed key for EventBridge to use, if you choose to use a customer managed key to encrypt this archive. The identifier can be the key Amazon Resource Name (ARN), KeyId, key alias, or key alias ARN.
* `retention_days` - (Optional) The maximum number of days to retain events in the new event archive. By default, it archives indefinitely.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the archive.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an EventBridge archive using the `name`. For example:

```terraform
import {
  to = aws_cloudwatch_event_archive.imported_event_archive.test
  id = "order-archive"
}
```

Using `terraform import`, import an EventBridge archive using the `name`. For example:

```console
% terraform import aws_cloudwatch_event_archive.imported_event_archive order-archive
```
