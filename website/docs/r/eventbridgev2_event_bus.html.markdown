---
subcategory: "EventBridge V2"
layout: "aws"
page_title: "AWS: aws_eventbridgev2_event_bus"
description: |-
  Manages an Amazon EventBridge V2 event bus.
---

# Resource: aws_eventbridgev2_event_bus

Manages an Amazon EventBridge V2 event bus. An event bus receives events and delivers them to subscribers, and can optionally store events for replay.

## Example Usage

### Basic Usage

```terraform
resource "aws_eventbridgev2_event_bus" "example" {
  name = "example"
}
```

### With Encryption and Storage Configuration

```terraform
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_kms_key" "example" {
  description             = "example"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "EnableRootAccess"
        Effect = "Allow"
        Principal = {
          AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "AllowEventBridgeServicePrincipal"
        Effect = "Allow"
        Principal = {
          Service = "events.amazonaws.com"
        }
        Action = [
          "kms:Decrypt",
          "kms:Encrypt",
          "kms:GenerateDataKeyWithoutPlaintext",
          "kms:ReEncryptFrom",
          "kms:ReEncryptTo",
        ]
        Resource = "*"
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
          StringLike = {
            "kms:EncryptionContext:aws:events:event-busv2:arn" = "arn:${data.aws_partition.current.partition}:events:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:event-busv2/example/*"
          }
        }
      },
    ]
  })
}

resource "aws_eventbridgev2_event_bus" "example" {
  name        = "example"
  description = "Example event bus"

  encryption_configuration = {
    kms_key_identifier = aws_kms_key.example.arn
  }

  storage_configuration = {
    retention_period_in_days = 1
  }
}
```

~> **Note:** When removing `encryption_configuration` or changing the KMS key, keep the previous key enabled until the update completes. EventBridge must be able to use the current key to update the event bus.

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the event bus. Must start with an alphanumeric character and contain only alphanumeric characters, periods (`.`), hyphens (`-`), and underscores (`_`). Changing this value creates a new resource.

The following arguments are optional:

* `description` - (Optional) Description of the event bus.
* `encryption_configuration` - (Optional) Encryption configuration for the event bus. If omitted, events are encrypted with an AWS owned key. Removing it reverts the event bus to an AWS owned key. See [`encryption_configuration` Block](#encryption_configuration-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `storage_configuration` - (Optional) Event storage configuration for the event bus. If omitted, the service default is applied. See [`storage_configuration` Block](#storage_configuration-block) below.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `encryption_configuration` Block

The `encryption_configuration` object supports the following:

* `kms_key_identifier` - (Optional) ARN of the symmetric AWS KMS customer managed key used to encrypt events. The key policy must grant EventBridge access to the key; see [Encryption at rest for custom event buses](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-custom-bus-encryption.html).

### `storage_configuration` Block

The `storage_configuration` object supports the following:

* `retention_period_in_days` - (Optional) Number of days events are retained on the event bus.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the event bus.
* `creation_time` - Time the event bus was created, in RFC3339 format.
* `last_modified_time` - Time the event bus was last modified, in RFC3339 format.
* `state` - Lifecycle state of the event bus.
* `state_reason` - Explanation of why the event bus is in its current state. Only set when the event bus is not `ACTIVE`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_eventbridgev2_event_bus.example
  identity = {
    "arn" = "arn:aws:events:us-east-1:123456789012:event-busv2/example/abcdefghij0123456789abcde"
  }
}

resource "aws_eventbridgev2_event_bus" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) ARN of the event bus.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EventBridge V2 event buses using the `arn`. For example:

```terraform
import {
  to = aws_eventbridgev2_event_bus.example
  id = "arn:aws:events:us-east-1:123456789012:event-busv2/example/abcdefghij0123456789abcde"
}
```

Using `terraform import`, import EventBridge V2 event buses using the `arn`. For example:

```console
% terraform import aws_eventbridgev2_event_bus.example arn:aws:events:us-east-1:123456789012:event-busv2/example/abcdefghij0123456789abcde
```
