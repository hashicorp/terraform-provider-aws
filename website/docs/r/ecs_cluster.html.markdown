---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_cluster"
description: |-
  Provides an ECS cluster.
---

# Resource: aws_ecs_cluster

Provides an ECS cluster.

## Example Usage

```terraform
resource "aws_ecs_cluster" "foo" {
  name = "white-hart"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}
```

### Execute Command Configuration with Override Logging

```terraform
resource "aws_kms_key" "example" {
  description             = "example"
  deletion_window_in_days = 7
}

resource "aws_cloudwatch_log_group" "example" {
  name = "example"
}

resource "aws_ecs_cluster" "test" {
  name = "example"

  configuration {
    execute_command_configuration {
      kms_key_id = aws_kms_key.example.arn
      logging    = "OVERRIDE"

      log_configuration {
        cloud_watch_encryption_enabled = true
        cloud_watch_log_group_name     = aws_cloudwatch_log_group.example.name
      }
    }
  }
}
```

### Fargate Ephemeral Storage Encryption with Customer-Managed KMS Key

```terraform
data "aws_caller_identity" "current" {}

resource "aws_kms_key" "example" {
  description             = "example"
  deletion_window_in_days = 7
}

resource "aws_kms_key_policy" "example" {
  key_id = aws_kms_key.example.id
  policy = jsonencode({
    Id = "ECSClusterFargatePolicy"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          "AWS" : "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow generate data key access for Fargate tasks."
        Effect = "Allow"
        Principal = {
          Service = "fargate.amazonaws.com"
        }
        Action = [
          "kms:GenerateDataKeyWithoutPlaintext"
        ]
        Condition = {
          StringEquals = {
            "kms:EncryptionContext:aws:ecs:clusterAccount" = [
              data.aws_caller_identity.current.account_id
            ]
            "kms:EncryptionContext:aws:ecs:clusterName" = [
              "example"
            ]
          }
        }
        Resource = "*"
      },
      {
        Sid    = "Allow grant creation permission for Fargate tasks."
        Effect = "Allow"
        Principal = {
          Service = "fargate.amazonaws.com"
        }
        Action = [
          "kms:CreateGrant"
        ]
        Condition = {
          StringEquals = {
            "kms:EncryptionContext:aws:ecs:clusterAccount" = [
              data.aws_caller_identity.current.account_id
            ]
            "kms:EncryptionContext:aws:ecs:clusterName" = [
              "example"
            ]
          }
          "ForAllValues:StringEquals" = {
            "kms:GrantOperations" = [
              "Decrypt"
            ]
          }
        }
        Resource = "*"
      }
    ]
    Version = "2012-10-17"
  })
}

resource "aws_ecs_cluster" "test" {
  name = "example"

  configuration {
    managed_storage_configuration {
      fargate_ephemeral_storage_kms_key_id = aws_kms_key.example.id
    }
  }
  depends_on = [
    aws_kms_key_policy.example
  ]
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the cluster (up to 255 letters, numbers, hyphens, and underscores)

The following arguments are optional:

* `configuration` - (Optional) Execute command configuration for the cluster. See [`configueration` Block](#configuration-block) for details.
* `service_connect_defaults` - (Optional) Default Service Connect namespace. See [`service_connect_defaults` Block](#service_connect_defaults-block) for details.
* `setting` - (Optional) Configuration block(s) with cluster settings. For example, this can be used to enable CloudWatch Container Insights for a cluster. See [`setting` Block](#setting-block) for details.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `configuration` Block

The `configuration` configuration block supports the following arguments:

* `execute_command_configuration` - (Optional) Details of the execute command configuration. See [`execute_command_configuration` Block](#execute_command_configuration-block) for details.
* `managed_storage_configuration` - (Optional) Details of the managed storage configuration. See [`managed_storage_configuration` Block](#managed_storage_configuration-block) for details.

### `execute_command_configuration` Block

The `execute_command_configuration` configuration block supports the following arguments:

* `kms_key_id` - (Optional) AWS Key Management Service key ID to encrypt the data between the local client and the container.
* `log_configuration` - (Optional) Log configuration for the results of the execute command actions. Required when `logging` is `OVERRIDE`. See [`log_configuration` Block](#log_configuration-block) for details.
* `logging` - (Optional) Log setting to use for redirecting logs for your execute command results. Valid values: `NONE`, `DEFAULT`, `OVERRIDE`.

#### `log_configuration` Block

The `log_configuration` configuration block supports the following arguments:

* `cloud_watch_encryption_enabled` - (Optional) Whether to enable encryption on the CloudWatch logs. If not specified, encryption will be disabled.
* `cloud_watch_log_group_name` - (Optional) The name of the CloudWatch log group to send logs to.
* `s3_bucket_name` - (Optional) Name of the S3 bucket to send logs to.
* `s3_bucket_encryption_enabled` - (Optional) Whether to enable encryption on the logs sent to S3. If not specified, encryption will be disabled.
* `s3_key_prefix` - (Optional) Optional folder in the S3 bucket to place logs in.

### `managed_storage_configuration` Block

The `managed_storage_configuration` configuration block supports the following arguments:

* `fargate_ephemeral_storage_kms_key_id` - (Optional) AWS Key Management Service key ID for the Fargate ephemeral storage.
* `kms_key_id` - (Optional) AWS Key Management Service key ID to encrypt the managed storage.

### `service_connect_defaults` Block

The `service_connect_defaults` configuration block supports the following arguments:

* `namespace` - (Required) ARN of the [`aws_service_discovery_http_namespace`](/docs/providers/aws/r/service_discovery_http_namespace.html) that's used when you create a service and don't specify a Service Connect configuration.

### `setting` Block

The `setting` configuration block supports the following arguments:

* `name` - (Required) Name of the setting to manage. Valid values: `containerInsights`.
* `value` -  (Required) Value to assign to the setting. Valid values: `enabled`, `disabled`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN that identifies the cluster.
* `id` - ARN that identifies the cluster.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ECS clusters using the cluster name. For example:

```terraform
import {
  to = aws_ecs_cluster.stateless
  id = "stateless-app"
}
```

Using `terraform import`, import ECS clusters using the cluster name. For example:

```console
% terraform import aws_ecs_cluster.stateless stateless-app
```
